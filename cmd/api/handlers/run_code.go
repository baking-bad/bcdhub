package handlers

import (
	"net/http"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/operations"
	"github.com/baking-bad/bcdhub/internal/parsers/storage"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// RunOperation -
func (ctx *Context) RunOperation(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusNotFound) {
		return
	}
	var reqRunOp runOperationRequest
	if err := c.BindJSON(&reqRunOp); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	state, err := ctx.CachedCurrentBlock(req.Network)
	if ctx.handleError(c, err, 0) {
		return
	}

	parameters, err := ctx.buildParametersForExecution(req.Network, req.Address, state.Protocol, reqRunOp.Name, reqRunOp.Data)
	if ctx.handleError(c, err, 0) {
		return
	}

	rpc, err := ctx.GetRPC(req.Network)
	if ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	counter, err := rpc.GetCounter(reqRunOp.Source)
	if ctx.handleError(c, err, 0) {
		return
	}

	protocol, err := ctx.Protocols.Get(req.Network, "", -1)
	if ctx.handleError(c, err, 0) {
		return
	}

	params, err := json.Marshal(parameters)
	if ctx.handleError(c, err, 0) {
		return
	}

	response, err := rpc.RunOperation(
		state.ChainID,
		state.Hash,
		reqRunOp.Source,
		req.Address,
		0, // fee
		protocol.Constants.HardGasLimitPerOperation,
		protocol.Constants.HardStorageLimitPerOperation,
		counter+1,
		reqRunOp.Amount,
		params,
	)
	if ctx.handleError(c, err, 0) {
		return
	}

	header := noderpc.Header{
		Level:       state.Level,
		Protocol:    state.Protocol,
		Timestamp:   state.Timestamp,
		ChainID:     state.ChainID,
		Hash:        state.Hash,
		Predecessor: state.Predecessor,
	}

	parser := operations.NewGroup(operations.NewParseParams(
		rpc,
		ctx.Context,
		operations.WithConstants(*protocol.Constants),
		operations.WithHead(header),
		operations.WithShareDirectory(ctx.SharePath),
		operations.WithNetwork(req.Network),
	))

	parsedModels, err := parser.Parse(response)
	if ctx.handleError(c, err, 0) {
		return
	}

	resp := make([]Operation, len(parsedModels.Operations))
	for i := range parsedModels.Operations {
		bmd := make([]bigmapdiff.BigMapDiff, 0)
		for j := range parsedModels.BigMapDiffs {
			if parsedModels.BigMapDiffs[j].OperationHash == parsedModels.Operations[i].Hash &&
				parsedModels.BigMapDiffs[j].OperationCounter == parsedModels.Operations[i].Counter &&
				helpers.IsInt64PointersEqual(parsedModels.BigMapDiffs[j].OperationNonce, parsedModels.Operations[i].Nonce) {
				bmd = append(bmd, *parsedModels.BigMapDiffs[j])
			}
		}
		op, err := ctx.prepareOperation(*parsedModels.Operations[i], bmd, true)
		if ctx.handleError(c, err, 0) {
			return
		}
		resp[i] = op
	}

	c.JSON(http.StatusOK, resp)
}

// RunCode godoc
// @Summary Execute entrypoint with passed arguments
// @Description Execute entrypoint with passed arguments
// @Tags contract
// @ID run-code
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param body body runCodeRequest true "Request body"
// @Accept json
// @Produce json
// @Success 200 {array} Operation
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/entrypoints/trace [post]
func (ctx *Context) RunCode(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusNotFound) {
		return
	}
	var reqRunCode runCodeRequest
	if err := c.BindJSON(&reqRunCode); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	state, err := ctx.CachedCurrentBlock(req.Network)
	if ctx.handleError(c, err, 0) {
		return
	}

	scriptBytes, err := ctx.getScriptBytes(req.Address, req.Network, state.Protocol)
	if ctx.handleError(c, err, 0) {
		return
	}

	input, err := ctx.buildParametersForExecution(req.Network, req.Address, state.Protocol, reqRunCode.Name, reqRunCode.Data)
	if ctx.handleError(c, err, 0) {
		return
	}

	rpc, err := ctx.GetRPC(req.Network)
	if ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	storage, err := rpc.GetScriptStorageRaw(req.Address, 0)
	if ctx.handleError(c, err, 0) {
		return
	}

	main := Operation{
		IndexedTime: time.Now().UTC().UnixNano(),
		Protocol:    state.Protocol,
		Network:     req.Network,
		Timestamp:   time.Now().UTC(),
		Source:      reqRunCode.Source,
		Destination: req.Address,
		GasLimit:    reqRunCode.GasLimit,
		Amount:      reqRunCode.Amount,
		Kind:        consts.Transaction,
		Level:       state.Level,
		Status:      consts.Applied,
		Entrypoint:  input.Entrypoint,
	}

	response, err := rpc.RunCode(scriptBytes, storage, input.Value, state.ChainID, reqRunCode.Source, reqRunCode.Sender, input.Entrypoint, state.Protocol, reqRunCode.Amount, reqRunCode.GasLimit)
	if err != nil {
		var e noderpc.InvalidNodeResponse
		if errors.As(err, &e) {
			main.Status = consts.Failed
			errs, err := tezerrors.ParseArray(e.Raw)
			if err != nil {
				ctx.handleError(c, err, 0)
				return
			}
			main.Errors = errs
			if err := formatErrors(main.Errors, &main); err != nil {
				ctx.handleError(c, err, 0)
				return
			}
			c.JSON(http.StatusOK, []Operation{main})
			return
		}
		ctx.handleError(c, err, 0)
		return
	}

	script, err := ast.NewScript(scriptBytes)
	if ctx.handleError(c, err, 0) {
		return
	}
	if err := setParatemetersWithType(input, script, &main); ctx.handleError(c, err, 0) {
		return
	}
	if err := ctx.setSimulateStorageDiff(response, script, &main); ctx.handleError(c, err, 0) {
		return
	}
	operations, err := ctx.parseAppliedRunCode(response, script, &main)
	if ctx.handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, operations)
}

func (ctx *Context) parseAppliedRunCode(response noderpc.RunCodeResponse, script *ast.Script, main *Operation) ([]Operation, error) {
	operations := []Operation{*main}

	for i := range response.Operations {
		var op Operation
		op.Kind = response.Operations[i].Kind
		op.Amount = *response.Operations[i].Amount
		op.Source = response.Operations[i].Source
		op.Destination = *response.Operations[i].Destination
		op.Status = consts.Applied
		op.Network = main.Network
		op.Timestamp = main.Timestamp
		op.Protocol = main.Protocol
		op.Level = main.Level
		op.Internal = true

		var s *ast.Script
		if op.Destination == main.Destination {
			s = script
		} else {
			var err error
			s, err = ctx.getScript(op.Destination, op.Network, op.Protocol)
			if err != nil {
				return nil, err
			}
		}

		if err := setParameters(response.Operations[i].Parameters, s, &op); err != nil {
			return nil, err
		}
		if err := ctx.setSimulateStorageDiff(response, script, &op); err != nil {
			return nil, err
		}
		operations = append(operations, op)
	}
	return operations, nil
}

func (ctx *Context) parseBigMapDiffs(response noderpc.RunCodeResponse, script *ast.Script, operation *Operation) ([]bigmapdiff.BigMapDiff, error) {
	model := operation.ToModel()
	model.AST = script

	rpc, err := ctx.GetRPC(operation.Network)
	if err != nil {
		return nil, err
	}

	parser, err := storage.MakeStorageParser(ctx.BigMapDiffs, rpc, operation.Protocol)
	if err != nil {
		return nil, err
	}

	nodeOperation := noderpc.Operation{
		Kind:         operation.Kind,
		Source:       operation.Source,
		Fee:          operation.Fee,
		Counter:      operation.Counter,
		GasLimit:     operation.GasLimit,
		StorageLimit: operation.StorageLimit,
		Amount:       &operation.Amount,
		Destination:  &operation.Destination,
		Delegate:     operation.Delegate,

		Result: &noderpc.OperationResult{
			Status:      consts.Applied,
			Storage:     response.Storage,
			BigMapDiffs: response.BigMapDiffs,
		},
	}

	rs := storage.RichStorage{Empty: true}
	switch operation.Kind {
	case consts.Transaction:
		rs, err = parser.ParseTransaction(nodeOperation, model)
	case consts.Origination:
		rs, err = parser.ParseOrigination(nodeOperation, model)
	}
	if err != nil {
		return nil, err
	}
	if rs.Empty {
		return nil, nil
	}
	bmd := make([]bigmapdiff.BigMapDiff, len(rs.Result.BigMapDiffs))
	for i := range rs.Result.BigMapDiffs {
		bmd[i] = *rs.Result.BigMapDiffs[i]
	}
	return bmd, nil
}

func (ctx *Context) setSimulateStorageDiff(response noderpc.RunCodeResponse, script *ast.Script, main *Operation) error {
	if len(response.Storage) == 0 || !bcd.IsContract(main.Destination) || main.Status != consts.Applied {
		return nil
	}
	bmd, err := ctx.parseBigMapDiffs(response, script, main)
	if err != nil {
		return err
	}
	storageType, err := script.StorageType()
	if err != nil {
		return err
	}
	storageDiff, err := ctx.getStorageDiff(bmd, main.Destination, response.Storage, storageType, main)
	if err != nil {
		return err
	}
	main.StorageDiff = storageDiff
	return nil
}
