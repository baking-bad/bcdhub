package handlers

import (
	"net/http"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/operations"
	"github.com/baking-bad/bcdhub/internal/parsers/storage"
	"github.com/baking-bad/bcdhub/internal/postgres"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// RunOperation -
func RunOperation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}
		var reqRunOp runOperationRequest
		if err := c.BindJSON(&reqRunOp); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		state, err := ctx.Blocks.Last()
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		predecessor, err := ctx.Blocks.Get(state.Level - 1)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		parameters, err := buildParametersForExecution(ctx, req.Address, state.Protocol.SymLink, reqRunOp.Name, reqRunOp.Data)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		counter, err := ctx.RPC.GetCounter(c, reqRunOp.Source)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		params, err := json.Marshal(parameters)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		response, err := ctx.RPC.RunOperationLight(
			c,
			state.Protocol.ChainID,
			state.Hash,
			reqRunOp.Source,
			req.Address,
			0, // fee
			state.Protocol.Constants.HardGasLimitPerOperation,
			state.Protocol.Constants.HardStorageLimitPerOperation,
			counter+1,
			reqRunOp.Amount,
			params,
		)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		header := noderpc.Header{
			Level:       state.Level,
			Protocol:    state.Protocol.Hash,
			Timestamp:   state.Timestamp,
			ChainID:     state.Protocol.ChainID,
			Hash:        state.Hash,
			Predecessor: predecessor.Hash,
		}

		parserParams, err := operations.NewParseParams(
			ctx,
			operations.WithProtocol(&state.Protocol),
			operations.WithHead(header),
		)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		parser := operations.NewGroup(parserParams)

		store := postgres.NewStore(ctx.StorageDB.DB)
		if err := parser.Parse(response, store); handleError(c, ctx.Storage, err, 0) {
			return
		}
		operations := store.ListOperations()

		resp := make([]Operation, len(operations))
		for i := range operations {
			bmd := make([]bigmapdiff.BigMapDiff, 0)
			for j := range operations[i].BigMapDiffs {
				bmd = append(bmd, *operations[i].BigMapDiffs[j])
			}
			op, err := prepareOperation(ctx, *operations[i], bmd, true)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			resp[i] = op
		}

		c.SecureJSON(http.StatusOK, resp)
	}
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
func RunCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}
		var reqRunCode runCodeRequest
		if err := c.BindJSON(&reqRunCode); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		state, err := ctx.Cache.CurrentBlock()
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		scriptBytes, err := getScriptBytes(ctx, req.Address, state.Protocol.SymLink)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		input, err := buildParametersForExecution(ctx, req.Address, state.Protocol.SymLink, reqRunCode.Name, reqRunCode.Data)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		storage, err := ctx.RPC.GetScriptStorageRaw(c, req.Address, 0)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		main := Operation{
			IndexedTime: time.Now().UTC().UnixNano(),
			Protocol:    state.Protocol.Hash,
			Network:     req.Network,
			Timestamp:   time.Now().UTC(),
			Source:      reqRunCode.Source,
			Destination: req.Address,
			GasLimit:    reqRunCode.GasLimit,
			Amount:      reqRunCode.Amount,
			Kind:        types.OperationKindTransaction.String(),
			Level:       state.Level,
			Status:      consts.Applied,
			Entrypoint:  input.Entrypoint,
		}

		response, err := ctx.RPC.RunCode(c, scriptBytes, storage, input.Value, state.Protocol.ChainID, reqRunCode.Source, reqRunCode.Sender, input.Entrypoint, state.Protocol.Hash, reqRunCode.Amount, reqRunCode.GasLimit)
		if err != nil {
			var e noderpc.InvalidNodeResponse
			if errors.As(err, &e) {
				main.Status = consts.Failed
				errs, err := tezerrors.ParseArray(e.Raw)
				if err != nil {
					handleError(c, ctx.Storage, e, 0)
					return
				}
				main.Errors = errs
				if err := formatErrors(main.Errors, &main); err != nil {
					handleError(c, ctx.Storage, err, 0)
					return
				}
				c.SecureJSON(http.StatusOK, []Operation{main})
				return
			}
			handleError(c, ctx.Storage, err, 0)
			return
		}

		script, err := ast.NewScript(scriptBytes)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		if err := setParatemetersWithType(input, script, &main); handleError(c, ctx.Storage, err, 0) {
			return
		}
		if err := setSimulateStorageDiff(c, ctx, response, state.Protocol, script, &main); handleError(c, ctx.Storage, err, 0) {
			return
		}
		operations, err := parseAppliedRunCode(c, ctx, response, script, &main, state.Protocol)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, operations)
	}
}

func parseAppliedRunCode(c *gin.Context, ctx *config.Context, response noderpc.RunCodeResponse, script *ast.Script, main *Operation, proto protocol.Protocol) ([]Operation, error) {
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
			s, err = getScript(ctx, op.Destination, proto.SymLink)
			if err != nil {
				return nil, err
			}
		}

		if err := setParameters(response.Operations[i].Parameters, s, &op); err != nil {
			return nil, err
		}
		if err := setSimulateStorageDiff(c, ctx, response, proto, script, &op); err != nil {
			return nil, err
		}
		operations = append(operations, op)
	}
	return operations, nil
}

func parseBigMapDiffs(c *gin.Context, ctx *config.Context, response noderpc.RunCodeResponse, script *ast.Script, operation *Operation, proto protocol.Protocol) ([]bigmapdiff.BigMapDiff, error) {
	model := operation.ToModel()
	model.AST = script

	model.ProtocolID = proto.ID

	parser, err := storage.MakeStorageParser(ctx.BigMapDiffs, ctx.RPC, operation.Protocol)
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

	store := postgres.NewStore(ctx.StorageDB.DB)
	switch operation.Kind {
	case types.OperationKindTransaction.String():
		err = parser.ParseTransaction(c, nodeOperation, &model, store)
	case types.OperationKindOrigination.String(), types.OperationKindOriginationNew.String():
		_, err = parser.ParseOrigination(c, nodeOperation, &model, store)
	}
	if err != nil {
		return nil, err
	}
	bmd := make([]bigmapdiff.BigMapDiff, len(model.BigMapDiffs))
	for i := range model.BigMapDiffs {
		bmd[i] = *model.BigMapDiffs[i]
	}
	return bmd, nil
}

func setSimulateStorageDiff(c *gin.Context, ctx *config.Context, response noderpc.RunCodeResponse, proto protocol.Protocol, script *ast.Script, main *Operation) error {
	if len(response.Storage) == 0 || !bcd.IsContract(main.Destination) || main.Status != consts.Applied {
		return nil
	}
	bmd, err := parseBigMapDiffs(c, ctx, response, script, main, proto)
	if err != nil {
		return err
	}
	storageType, err := script.StorageType()
	if err != nil {
		return err
	}

	destination, err := ctx.Accounts.Get(main.Destination)
	if err != nil {
		return err
	}

	storageDiff, err := getStorageDiff(ctx, destination.ID, bmd, response.Storage, storageType, main)
	if err != nil {
		return err
	}
	main.StorageDiff = storageDiff
	return nil
}
