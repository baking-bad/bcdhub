package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	formattererror "github.com/baking-bad/bcdhub/internal/bcd/formatter/error"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/parsers/storage"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// GetContractOperations godoc
// @Summary Get contract operations
// @Description Get contract operations
// @Tags contract
// @ID get-contract-operations
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param last_id query string false "Last operation ID"
// @Param from query integer false "Timestamp"
// @Param to query integer false "Timestamp"
// @Param size query integer false "Expected OPG count" mininum(1)
// @Param status query string false "Comma-separated operations statuses"
// @Param entrypoints query string false "Comma-separated called entrypoints list"
// @Param with_storage_diff query bool false "Include storage diff to operations or not"
// @Accept  json
// @Produce  json
// @Success 200 {object} OperationResponse
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/operations [get]
func (ctx *Context) GetContractOperations(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	var filtersReq operationsRequest
	if err := c.BindQuery(&filtersReq); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	filters := prepareFilters(filtersReq)
	ops, err := ctx.Operations.GetByContract(req.Network, req.Address, filtersReq.Size, filters)
	if ctx.handleError(c, err, 0) {
		return
	}

	resp, err := ctx.PrepareOperations(ops.Operations, filtersReq.WithStorageDiff)
	if ctx.handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, OperationResponse{
		Operations: resp,
		LastID:     ops.LastID,
	})
}

// GetOperation godoc
// @Summary Get operation group
// @Description Get operation group by hash
// @Tags operations
// @ID get-opg
// @Param hash path string true "Operation group hash"  minlength(51) maxlength(51)
// @Param with_mempool query bool false "Search operation in mempool or not"
// @Accept  json
// @Produce  json
// @Success 200 {array} Operation
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/opg/{hash} [get]
func (ctx *Context) GetOperation(c *gin.Context) {
	var req OPGRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	var queryReq opgRequest
	if err := c.BindQuery(&queryReq); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	op, err := ctx.Operations.Get(
		map[string]interface{}{
			"hash": req.Hash,
		},
		0,
		true,
	)
	if !ctx.Storage.IsRecordNotFound(err) && ctx.handleError(c, err, 0) {
		return
	}

	if len(op) == 0 {
		opg := make([]Operation, 0)

		if queryReq.WithMempool {
			operation := ctx.getOperationFromMempool(req.Hash)
			if operation != nil {
				opg = append(opg, *operation)
			}
		}

		c.JSON(http.StatusOK, opg)
		return
	}

	resp, err := ctx.PrepareOperations(op, true)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetOperationErrorLocation godoc
// @Summary Get code line where operation failed
// @DescriptionGet code line where operation failed
// @Tags operations
// @ID get-operation-error-location
// @Param id path string true "Internal BCD operation ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} GetErrorLocationResponse
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/operation/{id}/error_location [get]
func (ctx *Context) GetOperationErrorLocation(c *gin.Context) {
	var req getOperationByIDRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	operation := operation.Operation{ID: req.ID}
	if err := ctx.Storage.GetByID(&operation); ctx.handleError(c, err, 0) {
		return
	}

	if !tezerrors.HasScriptRejectedError(operation.Errors) {
		ctx.handleError(c, errors.Errorf("No reject script error in operation"), http.StatusBadRequest)
		return
	}

	response, err := ctx.getErrorLocation(operation, 2)
	if ctx.handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, response)
}

func (ctx *Context) getOperationFromMempool(hash string) *Operation {
	var wg sync.WaitGroup
	var opCh = make(chan *Operation, len(ctx.TzKTServices))

	defer close(opCh)

	for network := range ctx.TzKTServices {
		wg.Add(1)
		go ctx.getOperation(network, hash, opCh, &wg)
	}

	wg.Wait()

	for i := 0; i < len(ctx.TzKTServices); i++ {
		if op := <-opCh; op != nil {
			return op
		}
	}

	return nil
}

func (ctx *Context) getOperation(network, hash string, ops chan<- *Operation, wg *sync.WaitGroup) {
	defer wg.Done()

	api, err := ctx.GetTzKTService(network)
	if err != nil {
		ops <- nil
		return
	}

	res, err := api.GetMempool(hash)
	if err != nil {
		ops <- nil
		return
	}

	if len(res) == 0 {
		ops <- nil
		return
	}

	ops <- ctx.prepareMempoolOperation(res[0], network, string(res[0].Raw))
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

func formatErrors(errs []*tezerrors.Error, op *Operation) error {
	for i := range errs {
		if err := errs[i].Format(); err != nil {
			return err
		}
	}
	op.Errors = errs
	return nil
}

func (ctx *Context) prepareOperation(operation operation.Operation, bmd []bigmapdiff.BigMapDiff, withStorageDiff bool) (Operation, error) {
	log.Println("-----------------------------------------------")
	log.Println(operation.Destination)
	var op Operation
	op.FromModel(operation)

	var result OperationResult
	result.FromModel(operation.Result)
	op.Result = &result

	if err := formatErrors(operation.Errors, &op); err != nil {
		return op, err
	}
	if withStorageDiff {
		if operation.DeffatedStorage != "" && (operation.IsCall() || operation.IsOrigination()) && operation.IsApplied() {
			if err := ctx.setStorageDiff(op.Destination, operation.DeffatedStorage, &op, bmd); err != nil {
				return op, err
			}
		}
	}

	if !operation.IsTransaction() {
		return op, nil
	}

	if bcd.IsContract(op.Destination) && !tezerrors.HasParametersError(op.Errors) {
		if err := ctx.setParameters(operation.Parameters, &op); err != nil {
			return op, err
		}
	}

	return op, nil
}

// PrepareOperations -
func (ctx *Context) PrepareOperations(ops []operation.Operation, withStorageDiff bool) ([]Operation, error) {
	resp := make([]Operation, len(ops))
	for i := 0; i < len(ops); i++ {
		var bmd []bigmapdiff.BigMapDiff
		var err error

		if withStorageDiff {
			bmd, err = ctx.BigMapDiffs.GetUniqueByOperationID(ops[i].ID)
			if err != nil {
				return nil, err
			}
		}

		op, err := ctx.prepareOperation(ops[i], bmd, withStorageDiff)
		if err != nil {
			return nil, err
		}
		resp[i] = op
	}
	return resp, nil
}

func (ctx *Context) setParameters(parameters string, op *Operation) error {
	metadata, err := meta.GetSchema(ctx.Schema, op.Destination, consts.PARAMETER, op.Protocol)
	if err != nil {
		return nil
	}

	params := gjson.Parse(parameters)
	op.Parameters, err = newmiguel.ParameterToMiguel(params, metadata)
	if err != nil {
		if !tezerrors.HasGasExhaustedError(op.Errors) {
			helpers.CatchErrorSentry(err)
			return err
		}
	}
	return nil
}

func (ctx *Context) setStorageDiff(address, storage string, op *Operation, bmd []bigmapdiff.BigMapDiff) error {
	metadata, err := meta.GetContractSchema(ctx.Schema, address)
	if err != nil {
		return err
	}
	storageDiff, err := ctx.getStorageDiff(bmd, address, storage, metadata, false, op)
	if err != nil {
		return err
	}
	op.StorageDiff = storageDiff
	return nil
}

func (ctx *Context) getStorageDiff(bmd []bigmapdiff.BigMapDiff, address, storage string, metadata *meta.ContractSchema, isSimulating bool, op *Operation) (interface{}, error) {
	var prevStorage *newmiguel.Node
	var prevDeffatedStorage string
	prev, err := ctx.Operations.Last(op.Network, address, op.IndexedTime)
	if err == nil {
		prevBmd, err := ctx.getPrevBmd(bmd, op.IndexedTime, op.Destination)
		if err != nil {
			return nil, err
		}

		var exDeffatedStorage string
		exOp, err := ctx.Operations.Last(address, prev.Network, prev.IndexedTime)
		if err == nil {
			exDeffatedStorage = exOp.DeffatedStorage
		} else if !ctx.Storage.IsRecordNotFound(err) {
			return nil, err
		}

		prevStorage, err = getEnrichStorageMiguel(prevBmd, prev.Protocol, prev.DeffatedStorage, exDeffatedStorage, metadata, isSimulating)
		if err != nil {
			return nil, err
		}
		prevDeffatedStorage = prev.DeffatedStorage
	} else {
		if !ctx.Storage.IsRecordNotFound(err) {
			return nil, err
		}
		prevStorage = nil
	}

	currentStorage, err := getEnrichStorageMiguel(bmd, op.Protocol, storage, prevDeffatedStorage, metadata, isSimulating)
	if err != nil {
		return nil, err
	}
	if currentStorage == nil {
		return nil, nil
	}

	currentStorage.Diff(prevStorage)
	b, _ := json.Marshal(currentStorage)
	logger.Debug(string(b))
	return currentStorage, nil
}

func getEnrichStorageMiguel(bmd []bigmapdiff.BigMapDiff, protocol, storage, prevStorage string, metadata *meta.ContractSchema, isSimulating bool) (*newmiguel.Node, error) {
	store, err := enrichStorage(storage, prevStorage, bmd, protocol, false, isSimulating)
	if err != nil {
		return nil, err
	}
	logger.Debug(store.Raw)
	storageMetadata, err := metadata.Get(consts.STORAGE, protocol)
	if err != nil {
		return nil, err
	}
	return newmiguel.MichelineToMiguel(store, storageMetadata)
}

func enrichStorage(s, prevStorage string, bmd []bigmapdiff.BigMapDiff, protocol string, skipEmpty, isSimulating bool) (gjson.Result, error) {
	if len(bmd) == 0 {
		return gjson.Parse(s), nil
	}

	return storage.Enrich(s, bmd, skipEmpty, true)
}

func (ctx *Context) getPrevBmd(bmd []bigmapdiff.BigMapDiff, indexedTime int64, address string) ([]bigmapdiff.BigMapDiff, error) {
	if len(bmd) == 0 {
		return nil, nil
	}
	return ctx.BigMapDiffs.Previous(bmd, indexedTime, address)
}

func (ctx *Context) getErrorLocation(operation operation.Operation, window int) (GetErrorLocationResponse, error) {
	rpc, err := ctx.GetRPC(operation.Network)
	if err != nil {
		return GetErrorLocationResponse{}, err
	}
	code, err := contractparser.GetContract(rpc, operation.Destination, operation.Network, operation.Protocol, ctx.SharePath, 0)
	if err != nil {
		return GetErrorLocationResponse{}, err
	}
	opErr := tezerrors.First(operation.Errors, consts.ScriptRejectedError)
	if opErr == nil {
		return GetErrorLocationResponse{}, errors.Errorf("Can't find script rejected error")
	}
	defaultError, ok := opErr.IError.(*tezerrors.DefaultError)
	if !ok {
		return GetErrorLocationResponse{}, errors.Errorf("Invalid error type: %T", opErr)
	}

	location := int(defaultError.Location)
	sections := code.Get("code")
	row, sCol, eCol, err := formattererror.LocateContractError(sections, location)
	if err != nil {
		return GetErrorLocationResponse{}, err
	}

	michelson, err := formatter.MichelineToMichelson(sections, false, formatter.DefLineSize)
	if err != nil {
		return GetErrorLocationResponse{}, err
	}
	rows := strings.Split(michelson, "\n")
	start := helpers.MaxInt(0, row-window)
	end := helpers.MinInt(len(rows), row+window+1)

	rows = rows[start:end]
	return GetErrorLocationResponse{
		Text:        strings.Join(rows, "\n"),
		FailedRow:   row + 1,
		StartColumn: sCol,
		EndColumn:   eCol,
		FirstRow:    start + 1,
	}, nil
}
