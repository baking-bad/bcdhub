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
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	modelTypes "github.com/baking-bad/bcdhub/internal/models/types"
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
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/operations [get]
func (ctx *Context) GetContractOperations(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusNotFound) {
		return
	}

	var filtersReq operationsRequest
	if err := c.BindQuery(&filtersReq); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	account, err := ctx.Accounts.Get(req.NetworkID(), req.Address)
	if ctx.handleError(c, err, http.StatusNotFound) {
		return
	}

	filters := prepareFilters(filtersReq)
	ops, err := ctx.Operations.GetByAccount(account, filtersReq.Size, filters)
	if ctx.handleError(c, err, 0) {
		return
	}

	resp, err := ctx.PrepareOperations(ops.Operations, filtersReq.WithStorageDiff)
	if ctx.handleError(c, err, 0) {
		return
	}
	c.SecureJSON(http.StatusOK, OperationResponse{
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
// @Success 204 {object} gin.H
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

		if len(op) == 0 {
			c.SecureJSON(http.StatusNoContent, gin.H{})
			return
		}

		c.SecureJSON(http.StatusOK, opg)
		return
	}

	resp, err := ctx.PrepareOperations(op, true)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.SecureJSON(http.StatusOK, resp)
}

// GetOperationErrorLocation godoc
// @Summary Get code line where operation failed
// @Description Get code line where operation failed
// @Tags operations
// @ID get-operation-error-location
// @Param id path integer true "Internal BCD operation ID"
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
	c.SecureJSON(http.StatusOK, response)
}

// GetOperationDiff godoc
// @Summary Get operation storage diff
// @DescriptionGet Get operation storage diff
// @Tags operations
// @ID get-operation-diff
// @Param id path integer true "Internal BCD operation ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} ast.MiguelNode
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/operation/{id}/diff [get]
func (ctx *Context) GetOperationDiff(c *gin.Context) {
	var req getOperationByIDRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	operation, err := ctx.Operations.GetByID(req.ID)
	if ctx.handleError(c, err, 0) {
		return
	}

	var result Operation
	result.FromModel(operation)

	if len(operation.DeffatedStorage) > 0 && (operation.IsCall() || operation.IsOrigination()) && operation.IsApplied() {
		proto, err := ctx.Cache.ProtocolByID(operation.Network, operation.ProtocolID)
		if ctx.handleError(c, err, 0) {
			return
		}
		result.Protocol = proto.Hash

		storageBytes, err := ctx.Contracts.ScriptPart(operation.Network, operation.Destination.Address, proto.SymLink, consts.STORAGE)
		if ctx.handleError(c, err, 0) {
			return
		}

		storageType, err := ast.NewTypedAstFromBytes(storageBytes)
		if ctx.handleError(c, err, 0) {
			return
		}

		bmd, err := ctx.BigMapDiffs.GetForOperation(operation.ID)
		if ctx.handleError(c, err, 0) {
			return
		}

		if err := ctx.setStorageDiff(operation.DestinationID, operation.DeffatedStorage, &result, bmd, storageType); ctx.handleError(c, err, 0) {
			return
		}
	}
	c.SecureJSON(http.StatusOK, result.StorageDiff)
}

func (ctx *Context) getOperationFromMempool(hash string) *Operation {
	var wg sync.WaitGroup
	var opCh = make(chan *Operation, len(ctx.MempoolServices))

	defer close(opCh)

	for network := range ctx.MempoolServices {
		wg.Add(1)
		go ctx.getOperation(network, hash, opCh, &wg)
	}

	wg.Wait()

	for i := 0; i < len(ctx.MempoolServices); i++ {
		if op := <-opCh; op != nil {
			return op
		}
	}

	return nil
}

func (ctx *Context) getOperation(network modelTypes.Network, hash string, ops chan<- *Operation, wg *sync.WaitGroup) {
	defer wg.Done()

	api, err := ctx.GetMempoolService(network)
	if err != nil {
		ops <- nil
		return
	}

	res, err := api.GetByHash(hash)
	if err != nil {
		ops <- nil
		return
	}

	switch {
	case len(res.Originations) > 0:
		ops <- ctx.prepareMempoolOrigination(network, res.Originations[0])
	case len(res.Transactions) > 0:
		ops <- ctx.prepareMempoolTransaction(network, res.Transactions[0])
	default:
		ops <- nil
	}
}

func prepareFilters(req operationsRequest) map[string]interface{} {
	filters := map[string]interface{}{}

	if req.LastID != "" {
		filters["last_id"] = req.LastID
	}

	if req.From > 0 {
		filters["from"] = req.From / 1000
	}

	if req.To > 0 {
		filters["to"] = req.To / 1000
	}

	if req.Status != "" {
		statusList := make([]modelTypes.OperationStatus, 0)
		for _, item := range strings.Split(req.Status, ",") {
			status := modelTypes.NewOperationStatus(item)
			statusList = append(statusList, status)
		}
		filters["status"] = statusList
	}

	if req.Entrypoints != "" {
		filters["entrypoints"] = strings.Split(req.Entrypoints, ",")
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

	op.SourceAlias = operation.Source.Alias
	op.DestinationAlias = operation.Destination.Alias

	proto, err := ctx.Cache.ProtocolByID(operation.Network, operation.ProtocolID)
	if err != nil {
		return op, err
	}
	op.Protocol = proto.Hash

	if bcd.IsContract(op.Destination) {
		if err := formatErrors(operation.Errors, &op); err != nil {
			return op, err
		}

		script, err := ctx.getScript(operation.Network, op.Destination, proto.SymLink)
		if err != nil {
			return op, err
		}

		if withStorageDiff {
			storageType, err := script.StorageType()
			if err != nil {
				return op, err
			}
			if len(operation.DeffatedStorage) > 0 && (operation.IsCall() || operation.IsOrigination()) && operation.IsApplied() {
				if err := ctx.setStorageDiff(operation.DestinationID, operation.DeffatedStorage, &op, bmd, storageType); err != nil {
					return op, err
				}
			}
		}

		if !operation.IsTransaction() {
			return op, nil
		}

		if operation.IsCall() && !tezerrors.HasParametersError(op.Errors) {
			if err := setParameters(operation.Parameters, script, &op); err != nil {
				return op, err
			}
		}
	}

	return op, nil
}

// PrepareOperations -
func (ctx *Context) PrepareOperations(ops []operation.Operation, withStorageDiff bool) ([]Operation, error) {
	ids := make([]int64, 0, len(ops))
	for i := 0; i < len(ops); i++ {
		ids = append(ids, ops[i].ID)
	}
	bmd := make(map[int64][]bigmapdiff.BigMapDiff)

	if withStorageDiff {
		data, err := ctx.BigMapDiffs.GetForOperations(ids...)
		if err != nil {
			return nil, err
		}
		for i := range data {
			id := data[i].OperationID
			if _, ok := bmd[id]; !ok {
				bmd[id] = []bigmapdiff.BigMapDiff{}
			}
			bmd[id] = append(bmd[id], data[i])
		}
	}

	resp := make([]Operation, len(ops))
	for i := 0; i < len(ops); i++ {
		diffs, ok := bmd[ops[i].ID]
		if !ok {
			diffs = nil
		}

		op, err := ctx.prepareOperation(ops[i], diffs, withStorageDiff)
		if err != nil {
			return nil, err
		}
		resp[i] = op
	}
	return resp, nil
}

func setParameters(data []byte, script *ast.Script, op *Operation) error {
	if len(data) == 0 {
		return nil
	}
	params := types.NewParameters(data)
	return setParatemetersWithType(params, script, op)
}

func setParatemetersWithType(params *types.Parameters, script *ast.Script, op *Operation) error {
	if params == nil {
		return errors.New("Empty parameters")
	}
	parameter, err := script.ParameterType()
	if err != nil {
		return err
	}
	tree, err := parameter.FromParameters(params)
	if err != nil {
		if tezerrors.HasGasExhaustedError(op.Errors) {
			return nil
		}
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

func (ctx *Context) setStorageDiff(destinationID int64, storage []byte, op *Operation, bmd []bigmapdiff.BigMapDiff, storageType *ast.TypedAst) error {
	storageDiff, err := ctx.getStorageDiff(destinationID, bmd, storage, storageType, op)
	if err != nil {
		return err
	}
	op.StorageDiff = storageDiff
	return nil
}

func (ctx *Context) getStorageDiff(destinationID int64, bmd []bigmapdiff.BigMapDiff, storage []byte, storageType *ast.TypedAst, op *Operation) (*ast.MiguelNode, error) {
	currentStorage := &ast.TypedAst{
		Nodes: []ast.Node{ast.Copy(storageType.Nodes[0])},
	}
	var prevStorage *ast.TypedAst

	prev, err := ctx.Operations.Last(
		map[string]interface{}{
			"operation.network": modelTypes.NewNetwork(op.Network),
			"destination_id":    destinationID,
			"status":            modelTypes.OperationStatusApplied,
		}, op.ID)
	if err == nil {
		prevStorage = &ast.TypedAst{
			Nodes: []ast.Node{ast.Copy(storageType.Nodes[0])},
		}

		prevBmd, err := ctx.BigMapDiffs.Previous(bmd)
		if err != nil {
			return nil, err
		}

		if len(prev.DeffatedStorage) > 0 {
			if len(prevBmd) > 0 {
				if err := prepareStorage(prevStorage, prev.DeffatedStorage, prevBmd); err != nil {
					return nil, err
				}
			} else {
				if err := prepareStorage(prevStorage, prev.DeffatedStorage, bmd); err != nil {
					return nil, err
				}
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
	return currentStorage.Diff(prevStorage)
}

func prepareStorage(storageType *ast.TypedAst, deffatedStorage []byte, bmd []bigmapdiff.BigMapDiff) error {
	var data ast.UntypedAST
	if err := json.Unmarshal(deffatedStorage, &data); err != nil {
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

func (ctx *Context) getErrorLocation(operation operation.Operation, window int) (GetErrorLocationResponse, error) {
	proto, err := ctx.Cache.ProtocolByID(operation.Network, operation.ProtocolID)
	if err != nil {
		return GetErrorLocationResponse{}, err
	}
	code, err := ctx.getScriptBytes(operation.Network, operation.Destination.Address, proto.SymLink)
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
