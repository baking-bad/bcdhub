package handlers

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/miguel"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/storage"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/helpers"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/r3labs/diff"
	"github.com/tidwall/gjson"
)

type offsetRequest struct {
	LastID string `form:"last_id"`
}

// GetContractOperations -
func (ctx *Context) GetContractOperations(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var offsetReq offsetRequest
	if err := c.BindQuery(&offsetReq); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	size := uint64(10)
	var lastID uint64
	if offsetReq.LastID != "" {
		l, err := strconv.ParseUint(offsetReq.LastID, 10, 64)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		lastID = l
	}
	ops, err := ctx.ES.GetContractOperations(req.Network, req.Address, lastID, size)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	resp, err := prepareOperations(ctx.ES, ops.Operations)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, OperationResponse{
		Operations: resp,
		LastID:     ops.LastID,
	})
}

// OPGRequest -
type OPGRequest struct {
	Hash string `uri:"hash"`
}

// GetOperation -
func (ctx *Context) GetOperation(c *gin.Context) {
	var req OPGRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	op, err := ctx.ES.GetOperationByHash(req.Hash)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	resp, err := prepareOperations(ctx.ES, op)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, resp)
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

		BalanceUpdates: operation.BalanceUpdates,
		Result:         operation.Result,
	}

	if operation.DeffatedStorage != "" && strings.HasPrefix(op.Destination, "KT") && op.Result.Status == "applied" {
		if err := setStorageDiff(es, op.Destination, op.Network, operation.DeffatedStorage, &op); err != nil {
			return op, err
		}
	}

	if op.Kind != consts.Transaction {
		return op, nil
	}

	if operation.Parameters != "" && strings.HasPrefix(op.Destination, "KT") && !contractparser.IsParametersError(op.Result.Errors) {
		metadata, err := meta.GetMetadata(es, op.Destination, op.Network, "parameter", op.Protocol)
		if err != nil {
			return op, err
		}

		params := gjson.Parse(operation.Parameters)

		op.Parameters, err = miguel.MichelineToMiguel(params, metadata)
		if err != nil {
			return op, err
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
	store, err := enrichStorage(storage, bmd, op.Protocol)
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
		prevStore, err := enrichStorage(prev.DeffatedStorage, prevBmd, op.Protocol)
		if err != nil {
			return err
		}

		prevStorage, err = miguel.MichelineToMiguel(prevStore, metadata)
		if err != nil {
			return err
		}
	} else {
		if !strings.Contains(err.Error(), "Operation not found") {
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

func enrichStorage(s string, bmd gjson.Result, protocol string) (gjson.Result, error) {
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
	return parser.Enrich(s, bmd)
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
				field.Elem().SetMapIndex(reflect.ValueOf("kind"), reflect.ValueOf("create"))
			} else if !field.IsNil() {
				field.Elem().SetMapIndex(reflect.ValueOf("kind"), reflect.ValueOf(typ))
			} else {
				field.Elem().SetMapIndex(reflect.ValueOf("kind"), reflect.ValueOf("delete"))
			}

			if from != nil && typ != "delete" {
				field.Elem().SetMapIndex(reflect.ValueOf("from"), reflect.ValueOf(from))
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
		value.SetMapIndex(reflect.ValueOf(path[0]), reflect.ValueOf(from))
		value.SetMapIndex(reflect.ValueOf("kind"), reflect.ValueOf("delete"))
		return nil
	}

	if len(path) == 1 {
		if value.Kind() == reflect.Map {
			if !value.IsValid() {
				value.SetMapIndex(reflect.ValueOf("kind"), reflect.ValueOf("create"))
			} else if !value.IsNil() {
				value.SetMapIndex(reflect.ValueOf("kind"), reflect.ValueOf(typ))
			} else {
				value.SetMapIndex(reflect.ValueOf("kind"), reflect.ValueOf("delete"))
			}

			if from != nil && typ != "delete" {
				value.SetMapIndex(reflect.ValueOf("from"), reflect.ValueOf(from))
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
