package handlers

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/miguel"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage"
	"github.com/baking-bad/bcdhub/internal/contractparser/unpack/rawbytes"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/r3labs/diff"
	"github.com/tidwall/gjson"
)

// GetContractOperations -
func (ctx *Context) GetContractOperations(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var filtersReq operationsRequest
	if err := c.BindQuery(&filtersReq); handleError(c, err, http.StatusBadRequest) {
		return
	}

	size := uint64(10)
	filters := prepareFilters(filtersReq)

	ops, err := ctx.ES.GetContractOperations(req.Network, req.Address, size, filters)
	if handleError(c, err, 0) {
		return
	}

	resp, err := prepareOperations(ctx.ES, ops.Operations)
	if handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, OperationResponse{
		Operations: resp,
		LastID:     ops.LastID,
	})
}

// GetOperation -
func (ctx *Context) GetOperation(c *gin.Context) {
	var req OPGRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	op, err := ctx.ES.GetOperationByHash(req.Hash)
	if handleError(c, err, 0) {
		return
	}

	resp, err := prepareOperations(ctx.ES, op)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

func prepareFilters(req operationsRequest) map[string]interface{} {
	filters := map[string]interface{}{}

	if req.LastID != "" {
		filters["last_id"] = req.LastID
	}

	if req.From > 0 {
		filters["from"] = req.From
	}

	if req.To > 0 {
		filters["to"] = req.To
	}

	if req.Status != "" {
		status := "'" + strings.Join(strings.Split(req.Status, ","), "','") + "'"
		filters["status"] = status
	}

	if req.Entrypoints != "" {
		entrypoints := "'" + strings.Join(strings.Split(req.Entrypoints, ","), "','") + "'"
		filters["entrypoints"] = entrypoints
	}
	return filters
}

func formatErrors(errs []cerrors.Error, op *Operation) error {
	for i := range errs {
		if errs[i].With != "" {
			text := gjson.Parse(errs[i].With)
			if text.Get("bytes").Exists() {
				data := text.Get("bytes").String()
				data = strings.TrimPrefix(data, "05")
				decodedString, err := rawbytes.ToMicheline(data)
				if err == nil {
					text = gjson.Parse(decodedString)
				}
			}
			errString, err := formatter.MichelineToMichelson(text, true, formatter.DefLineSize)
			if err != nil {
				return err
			}
			errs[i].With = errString
		}
	}
	op.Errors = errs
	return nil
}

func prepareOperation(es *elastic.Elastic, operation models.Operation) (Operation, error) {
	op := Operation{
		ID:        operation.ID,
		Protocol:  operation.Protocol,
		Hash:      operation.Hash,
		Network:   operation.Network,
		Internal:  operation.Internal,
		Timesatmp: operation.Timestamp,

		Level:         operation.Level,
		Kind:          operation.Kind,
		Source:        operation.Source,
		Fee:           operation.Fee,
		Counter:       operation.Counter,
		GasLimit:      operation.GasLimit,
		StorageLimit:  operation.StorageLimit,
		Amount:        operation.Amount,
		Destination:   operation.Destination,
		PublicKey:     operation.PublicKey,
		ManagerPubKey: operation.ManagerPubKey,
		Balance:       operation.Balance,
		Delegate:      operation.Delegate,
		Status:        operation.Status,
		Entrypoint:    operation.Entrypoint,

		BalanceUpdates: operation.BalanceUpdates,
		Result:         operation.Result,
	}

	if err := formatErrors(operation.Errors, &op); err != nil {
		log.Println(err)
		return op, err
	}

	if operation.DeffatedStorage != "" && strings.HasPrefix(op.Destination, "KT") && op.Status == "applied" {
		if err := setStorageDiff(es, op.Destination, op.Network, operation.DeffatedStorage, &op); err != nil {
			return op, err
		}
	}

	if op.Kind != consts.Transaction {
		return op, nil
	}

	if strings.HasPrefix(op.Destination, "KT") && !cerrors.HasParametersError(op.Errors) {
		metadata, err := meta.GetMetadata(es, op.Destination, op.Network, "parameter", op.Protocol)
		if err != nil {
			return op, nil
		}

		params := gjson.Parse(operation.Parameters)

		op.Parameters, err = miguel.MichelineToMiguel(params, metadata)
		if err != nil {
			if !cerrors.HasGasExhaustedError(op.Errors) {
				helpers.CatchErrorSentry(err)
				return op, err
			}
		}
	}

	return op, nil
}

func prepareOperations(es *elastic.Elastic, ops []models.Operation) ([]Operation, error) {
	resp := make([]Operation, len(ops))
	for i := 0; i < len(ops); i++ {
		op, err := prepareOperation(es, ops[i])
		if err != nil {
			return nil, err
		}
		resp[i] = op
	}
	return resp, nil
}

func setStorageDiff(es *elastic.Elastic, address, network string, storage string, op *Operation) error {
	metadata, err := meta.GetMetadata(es, address, network, "storage", op.Protocol)
	if err != nil {
		return err
	}
	bmd, err := es.GetBigMapDiffsByOperationID(op.ID)
	if err != nil {
		return err
	}
	store, err := enrichStorage(storage, bmd, op.Protocol, false)
	if err != nil {
		return err
	}
	currentStorage, err := miguel.MichelineToMiguel(store, metadata)
	if err != nil {
		return err
	}

	var prevStorage interface{}
	prev, err := es.GetPreviousOperation(address, op.Network, op.Level)
	if err == nil {
		var prevBmd gjson.Result
		if len(bmd.Array()) > 0 {
			prevBmd, err = getPrevBmd(es, bmd, op.Level, op.Destination)
			if err != nil {
				return err
			}
		}
		prevStore, err := enrichStorage(prev.DeffatedStorage, prevBmd, op.Protocol, false)
		if err != nil {
			return err
		}

		prevMetadata, err := meta.GetMetadata(es, address, network, "storage", prev.Protocol)
		if err != nil {
			return err
		}
		prevStorage, err = miguel.MichelineToMiguel(prevStore, prevMetadata)
		if err != nil {
			return err
		}
	} else {
		if !strings.Contains(err.Error(), "Unknown") {
			return err
		}

		if currentStorage == nil {
			return nil
		}
		switch reflect.ValueOf(currentStorage).Kind() {
		case reflect.Map:
			prevStorage = map[string]string{}
		case reflect.Slice:
			prevStorage = make([]interface{}, 0)
		default:
			return fmt.Errorf("Unsupported storage type: %T %v", currentStorage, currentStorage)
		}

	}

	changelog, err := diff.Diff(prevStorage, currentStorage)
	if err != nil {
		return err
	}

	if len(changelog) != 0 {
		if err := applyChanges(changelog, &currentStorage); err != nil {
			return err
		}
	}

	op.StorageDiff = currentStorage
	return nil
}

func enrichStorage(s string, bmd gjson.Result, protocol string, skipEmpty bool) (gjson.Result, error) {
	if len(bmd.Array()) == 0 {
		return gjson.Parse(s), nil
	}

	var parser storage.Parser
	if helpers.StringInArray(protocol, []string{
		consts.HashBabylon, consts.HashCarthage, consts.HashZeroBabylon,
	}) {
		parser = storage.NewBabylon(nil, nil)
	} else {
		parser = storage.NewAlpha(nil)
	}
	return parser.Enrich(s, bmd, skipEmpty)
}

func getPrevBmd(es *elastic.Elastic, bmd gjson.Result, level int64, address string) (gjson.Result, error) {
	keys := make([]string, 0)
	for _, b := range bmd.Array() {
		keys = append(keys, b.Get("key_hash").String())
	}

	return es.GetBigMapDiffsByKeyHash(keys, level, address)
}

func applyChanges(changelog diff.Changelog, v interface{}) (err error) {
	for _, c := range changelog {
		if err = applyChange(c.Path, c.From, c.Type, v); err != nil {
			return
		}
	}
	return
}

func applyChange(path []string, from interface{}, typ string, v interface{}) error {
	if len(path) == 0 {
		return nil
	}

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	idx, err := strconv.Atoi(path[0])
	if err == nil && val.Kind() == reflect.Slice {
		return applyChangeOnArray(path, from, typ, idx, v)
	}

	return applyChangeOnMap(path, from, typ, v)
}

func applyChangeOnArray(path []string, from interface{}, typ string, idx int, v interface{}) error {
	var elem, ptr, value reflect.Value
	ptr = reflect.ValueOf(v)
	if ptr.Kind() == reflect.Ptr {
		elem = ptr.Elem()
	}
	if elem.Kind() == reflect.Interface {
		value = elem.Elem()
	}

	var field reflect.Value
	var newField interface{}

	if value.Len() <= idx {
		field = reflect.ValueOf(&from)
		elem.Set(reflect.Append(value, reflect.ValueOf(from)))
	} else {
		field = value.Index(idx)
	}
	newField = field.Interface()

	if len(path) == 1 {
		if field.Kind() == reflect.Ptr {
			field = field.Elem()
		}
		if field.Elem().Kind() == reflect.Map {
			if !field.Elem().IsValid() {
				field.Elem().SetMapIndex(reflect.ValueOf("miguel_kind"), reflect.ValueOf("create"))
			} else if !field.IsNil() {
				field.Elem().SetMapIndex(reflect.ValueOf("miguel_kind"), reflect.ValueOf(typ))
			} else {
				field.Elem().SetMapIndex(reflect.ValueOf("miguel_kind"), reflect.ValueOf("delete"))
			}

			if from != nil && typ != "delete" {
				field.Elem().SetMapIndex(reflect.ValueOf("miguel_from"), reflect.ValueOf(from))
			}
		}
		return nil
	}
	if err := applyChange(path[1:], from, typ, &newField); err != nil {
		return err
	}
	field.Set(reflect.ValueOf(newField))
	return nil
}

func applyChangeOnMap(path []string, from interface{}, typ string, v interface{}) error {
	var elem, ptr, value reflect.Value
	ptr = reflect.ValueOf(v)
	if ptr.Kind() == reflect.Ptr {
		elem = ptr.Elem()
	}
	if elem.Kind() == reflect.Interface {
		value = elem.Elem()
	}

	var field reflect.Value
	for _, k := range value.MapKeys() {
		if k.String() != path[0] {
			continue
		}
		field = value.MapIndex(k)
	}

	if field.IsValid() && field.IsNil() {
		fromValue := reflect.ValueOf(from)
		fromValue.SetMapIndex(reflect.ValueOf("miguel_kind"), reflect.ValueOf("delete"))
		value.SetMapIndex(reflect.ValueOf(path[0]), fromValue)
		return nil
	}

	if field.IsValid() && field.CanInterface() {
		fieldValue := field.Interface()
		if fieldValue, ok := fieldValue.(map[string]interface{}); ok {
			fromValue := reflect.ValueOf(fieldValue)
			fromValue.SetMapIndex(reflect.ValueOf("miguel_kind"), reflect.ValueOf("create"))
			value.SetMapIndex(reflect.ValueOf(path[0]), fromValue)
			return nil
		}
	}

	if len(path) == 1 {

		if value.Kind() == reflect.Map {
			if from == nil || !value.IsValid() {
				value.SetMapIndex(reflect.ValueOf("miguel_kind"), reflect.ValueOf("create"))
			} else if !value.IsNil() {
				value.SetMapIndex(reflect.ValueOf("miguel_kind"), reflect.ValueOf(typ))
			} else {
				value.SetMapIndex(reflect.ValueOf("miguel_kind"), reflect.ValueOf("delete"))
			}

			if from != nil && typ != "delete" {
				value.SetMapIndex(reflect.ValueOf("miguel_from"), reflect.ValueOf(from))
			}
		}
		return nil
	}
	if !field.IsValid() {
		return fmt.Errorf("Invalid map field: %v", field)
	}
	newField := field.Interface()
	if err := applyChange(path[1:], from, typ, &newField); err != nil {
		return err
	}
	value.SetMapIndex(reflect.ValueOf(path[0]), reflect.ValueOf(newField))
	return nil
}
