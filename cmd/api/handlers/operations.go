package handlers

import (
	"net/http"
	"strings"
	"sync"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	formattererror "github.com/baking-bad/bcdhub/internal/bcd/formatter/error"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/fetch"
	"github.com/baking-bad/bcdhub/internal/helpers"
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
	var op Operation
	op.FromModel(operation)

	var result OperationResult
	result.FromModel(operation.Result)
	op.Result = &result

	if err := formatErrors(operation.Errors, &op); err != nil {
		return op, err
	}

	script, err := ctx.getScript(op.Destination, op.Network, op.Protocol)
	if err != nil {
		return op, err
	}

	if withStorageDiff {
		if operation.DeffatedStorage != "" && (operation.IsCall() || operation.IsOrigination()) && operation.IsApplied() {
			if err := ctx.setStorageDiff(op.Destination, operation.DeffatedStorage, &op, bmd, script); err != nil {
				return op, err
			}
		}
	}

	if !operation.IsTransaction() {
		return op, nil
	}

	if bcd.IsContract(op.Destination) && !tezerrors.HasParametersError(op.Errors) {
		if err := ctx.setParameters(operation.Parameters, script, &op); err != nil {
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

func (ctx *Context) setParameters(data string, script *ast.Script, op *Operation) error {
	parameter, err := script.ParameterType()
	if err != nil {
		return err
	}
	params := types.NewParameters([]byte(data))
	op.Entrypoint = params.Entrypoint

	tree, err := parameter.FromParameters(params)
	if err != nil {
		return err
	}

	op.Parameters, err = tree.ToMiguel()
	if err != nil {
		if !tezerrors.HasGasExhaustedError(op.Errors) {
			helpers.CatchErrorSentry(err)
			return err
		}
	}
	return nil
}

func (ctx *Context) setStorageDiff(address, storage string, op *Operation, bmd []bigmapdiff.BigMapDiff, script *ast.Script) error {
	storageType, err := script.StorageType()
	if err != nil {
		return err
	}
	storageDiff, err := ctx.getStorageDiff(bmd, address, storage, storageType, op)
	if err != nil {
		return err
	}
	op.StorageDiff = storageDiff
	return nil
}

func (ctx *Context) getStorageDiff(bmd []bigmapdiff.BigMapDiff, address, storage string, storageType *ast.TypedAst, op *Operation) (interface{}, error) {
	currentStorage := &ast.TypedAst{
		Nodes: []ast.Node{ast.Copy(storageType.Nodes[0])},
	}
	var prevStorage *ast.TypedAst

	prev, err := ctx.Operations.Last(op.Network, address, op.IndexedTime)
	if err == nil {
		prevStorage = &ast.TypedAst{
			Nodes: []ast.Node{ast.Copy(storageType.Nodes[0])},
		}

		prevBmd, err := ctx.getPrevBmd(bmd, op.IndexedTime, op.Destination)
		if err != nil {
			return nil, err
		}

		if prev.DeffatedStorage != "" {
			if err := prepareStorage(prevStorage, prev.DeffatedStorage, prevBmd); err != nil {
				return nil, err
			}
		}
	} else if !ctx.Storage.IsRecordNotFound(err) {
		return nil, err
	}

	if err := prepareStorage(currentStorage, storage, bmd); err != nil {
		return nil, err
	}
	if !currentStorage.IsSettled() {
		return nil, nil
	}
	if prevStorage == nil {
		return currentStorage.ToMiguel()
	}

	return currentStorage.Diff(prevStorage)
}

func prepareStorage(storageType *ast.TypedAst, deffatedStorage string, bmd []bigmapdiff.BigMapDiff) error {
	var data ast.UntypedAST
	if err := json.UnmarshalFromString(deffatedStorage, &data); err != nil {
		return err
	}

	if err := storageType.Settle(data); err != nil {
		return err
	}

	return getEnrichStorage(storageType, bmd)
}

func getEnrichStorage(storageType *ast.TypedAst, bmd []bigmapdiff.BigMapDiff) error {
	if len(bmd) == 0 {
		return nil
	}

	return storage.Enrich(storageType, bmd, false, true)
}

func (ctx *Context) getPrevBmd(bmd []bigmapdiff.BigMapDiff, indexedTime int64, address string) ([]bigmapdiff.BigMapDiff, error) {
	if len(bmd) == 0 {
		return nil, nil
	}
	return ctx.BigMapDiffs.Previous(bmd, indexedTime, address)
}

func (ctx *Context) getErrorLocation(operation operation.Operation, window int) (GetErrorLocationResponse, error) {
	code, err := fetch.Contract(operation.Destination, operation.Network, operation.Protocol, ctx.SharePath)
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
	sections := gjson.ParseBytes(code)
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
