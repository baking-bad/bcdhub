package handlers

import (
	"net/http"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
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

	filters := prepareFilters(filtersReq)
	ops, err := ctx.ES.GetContractOperations(req.Network, req.Address, filtersReq.Size, filters)
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

func formatErrors(errs []cerrors.IError, op *Operation) error {
	for i := range errs {
		if err := errs[i].Format(); err != nil {
			return err
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

		Level:            operation.Level,
		Kind:             operation.Kind,
		Source:           operation.Source,
		SourceAlias:      operation.SourceAlias,
		Fee:              operation.Fee,
		Counter:          operation.Counter,
		GasLimit:         operation.GasLimit,
		StorageLimit:     operation.StorageLimit,
		Amount:           operation.Amount,
		Destination:      operation.Destination,
		DestinationAlias: operation.DestinationAlias,
		PublicKey:        operation.PublicKey,
		ManagerPubKey:    operation.ManagerPubKey,
		Delegate:         operation.Delegate,
		Status:           operation.Status,
		Burned:           operation.Burned,
		Entrypoint:       operation.Entrypoint,
		IndexedTime:      operation.IndexedTime,

		BalanceUpdates: operation.BalanceUpdates,
		Result:         operation.Result,
	}

	if err := formatErrors(operation.Errors, &op); err != nil {
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
		metadata, err := meta.GetMetadata(es, op.Destination, consts.PARAMETER, op.Protocol)
		if err != nil {
			return op, nil
		}

		params := gjson.Parse(operation.Parameters)
		op.Parameters, err = newmiguel.ParameterToMiguel(params, metadata)
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
	metadata, err := meta.GetContractMetadata(es, address)
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
	storageMetadata, err := metadata.Get(consts.STORAGE, op.Protocol)
	if err != nil {
		return err
	}
	currentStorage, err := newmiguel.MichelineToMiguel(store, storageMetadata)
	if err != nil {
		return err
	}

	var prevStorage *newmiguel.Node
	prev, err := es.GetPreviousOperation(address, op.Network, op.IndexedTime)
	if err == nil {
		var prevBmd []models.BigMapDiff
		if len(bmd) > 0 {
			prevBmd, err = getPrevBmd(es, bmd, op.IndexedTime, op.Destination)
			if err != nil {
				return err
			}
		}
		prevStore, err := enrichStorage(prev.DeffatedStorage, prevBmd, op.Protocol, false)
		if err != nil {
			return err
		}

		prevMetadata, err := metadata.Get(consts.STORAGE, prev.Protocol)
		if err != nil {
			return err
		}
		prevStorage, err = newmiguel.MichelineToMiguel(prevStore, prevMetadata)
		if err != nil {
			return err
		}
	} else {
		if !strings.Contains(err.Error(), elastic.RecordNotFound) {
			return err
		}

		if currentStorage == nil {
			return nil
		}
		prevStorage = nil
	}

	currentStorage.Diff(prevStorage)
	op.StorageDiff = currentStorage
	return nil
}

func enrichStorage(s string, bmd []models.BigMapDiff, protocol string, skipEmpty bool) (gjson.Result, error) {
	if len(bmd) == 0 {
		return gjson.Parse(s), nil
	}

	parser, err := contractparser.MakeStorageParser(nil, protocol)
	if err != nil {
		return gjson.Result{}, err
	}

	return parser.Enrich(s, bmd, skipEmpty)
}

func getPrevBmd(es *elastic.Elastic, bmd []models.BigMapDiff, indexedTime int64, address string) ([]models.BigMapDiff, error) {
	keys := make([]string, 0)
	ptr := make([]int64, 0)
	for _, b := range bmd {
		keys = append(keys, b.KeyHash)
		ptr = append(ptr, b.Ptr)
	}

	return es.GetBigMapDiffsByKeyHashAndPtr(keys, ptr, indexedTime, address)
}
