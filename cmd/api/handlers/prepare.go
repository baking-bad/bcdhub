package handlers

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/encoding"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	formattererror "github.com/baking-bad/bcdhub/internal/bcd/formatter/error"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	modelTypes "github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/parsers/storage"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

func newOperationResponse(ctx context.Context, cfgCtx *config.Context, operation operation.Operation) (Operation, protocol.Protocol, error) {
	var response Operation
	response.FromModel(operation)
	response.Network = cfgCtx.Network.String()

	proto, err := cfgCtx.Cache.ProtocolByID(ctx, operation.ProtocolID)
	if err != nil {
		return response, proto, err
	}
	response.Protocol = proto.Hash

	err = formatErrors(operation.Errors, &response)
	return response, proto, err
}

func preparePayloadOperation(ctx context.Context, cfgCtx *config.Context, operation operation.Operation) (Operation, error) {
	response, _, err := newOperationResponse(ctx, cfgCtx, operation)
	if err != nil {
		return response, err
	}
	payloadType, err := ast.NewTypedAstFromBytes(operation.PayloadType)
	if err != nil {
		return response, err
	}
	if err := payloadType.SettleFromBytes(operation.Payload); err != nil {
		return response, err
	}
	payloadMiguel, err := payloadType.ToMiguel()
	if err != nil {
		return response, err
	}
	response.Payload = payloadMiguel
	return response, err
}

func prepareSrExecute(ctx context.Context, cfgCtx *config.Context, operation operation.Operation) (Operation, error) {
	response, _, err := newOperationResponse(ctx, cfgCtx, operation)
	if err != nil {
		return response, err
	}

	if len(operation.Payload) < 32 {
		return response, nil
	}

	commitment, err := encoding.EncodeBase58(operation.Payload[:32], []byte(encoding.PrefixSmartRollupCommitment))
	if err != nil {
		return response, err
	}
	response.Payload = []*ast.MiguelNode{
		{
			Prim: "pair",
			Type: "namedtuple",
			Children: []*ast.MiguelNode{
				{
					Prim:  "string",
					Type:  "string",
					Name:  getStringPointer("cemented_commitment"),
					Value: commitment,
				}, {
					Prim:  "bytes",
					Type:  "bytes",
					Name:  getStringPointer("output_proof"),
					Value: hex.EncodeToString(operation.Payload[32:]),
				},
			},
		},
	}

	return response, nil
}

func prepareTransaction(ctx context.Context, cfgCtx *config.Context, operation operation.Operation, withStorageDiff bool) (Operation, error) {
	response, proto, err := newOperationResponse(ctx, cfgCtx, operation)
	if err != nil {
		return response, err
	}

	if withStorageDiff && operation.CanHasStorageDiff() {
		if err := setFullStorage(ctx, cfgCtx, proto.SymLink, operation, &response); err != nil {
			return response, err
		}
	}

	if operation.IsCall() && !tezerrors.HasParametersError(response.Errors) {
		switch {
		case bcd.IsContract(response.Destination):
			parameterType, err := getParameterType(ctx, cfgCtx.Contracts, response.Destination, proto.SymLink)
			if err != nil {
				return response, err
			}
			if err := setParameters(operation.Parameters, parameterType, &response); err != nil {
				return response, err
			}
		case bcd.IsSmartRollupHash(response.Destination):
			rollup, err := cfgCtx.SmartRollups.Get(ctx, response.Destination)
			if err != nil {
				return response, err
			}
			tree, err := ast.NewTypedAstFromBytes(rollup.Type)
			if err != nil {
				return response, err
			}
			if err := setParameters(operation.Parameters, tree, &response); err != nil {
				return response, err
			}
		}
	}

	return response, nil
}

func prepareOrigination(ctx context.Context, cfgCtx *config.Context, operation operation.Operation, withStorageDiff bool) (Operation, error) {
	response, proto, err := newOperationResponse(ctx, cfgCtx, operation)
	if err != nil {
		return response, err
	}

	if !withStorageDiff {
		return response, nil
	}
	if !operation.CanHasStorageDiff() {
		return response, nil
	}

	err = setFullStorage(ctx, cfgCtx, proto.SymLink, operation, &response)
	return response, err
}

func setFullStorage(ctx context.Context, cfgCtx *config.Context, symLink string, operation operation.Operation, response *Operation) error {
	storageType, err := getStorageType(ctx, cfgCtx.Contracts, response.Destination, symLink)
	if err != nil {
		return err
	}

	var diffs []bigmapdiff.BigMapDiff

	if operation.BigMapDiffsCount > 0 {
		if len(operation.BigMapDiffs) == 0 {
			diffs, err = cfgCtx.BigMapDiffs.GetForOperation(ctx, operation.ID)
			if err != nil {
				return err
			}
		} else {
			diffs = make([]bigmapdiff.BigMapDiff, len(operation.BigMapDiffs))
			for i := range operation.BigMapDiffs {
				diffs[i] = *operation.BigMapDiffs[i]
			}
		}
	}

	return setStorageDiff(ctx, cfgCtx, operation.DestinationID, operation.DeffatedStorage, response, diffs, storageType)
}

func prepareOperation(ctx context.Context, cfgCtx *config.Context, operation operation.Operation, withStorageDiff bool) (Operation, error) {
	switch operation.Kind {
	case modelTypes.OperationKindEvent:
		return preparePayloadOperation(ctx, cfgCtx, operation)
	case modelTypes.OperationKindTransferTicket:
		return preparePayloadOperation(ctx, cfgCtx, operation)
	case modelTypes.OperationKindTransaction:
		return prepareTransaction(ctx, cfgCtx, operation, withStorageDiff)
	case modelTypes.OperationKindOrigination:
		return prepareOrigination(ctx, cfgCtx, operation, withStorageDiff)
	case modelTypes.OperationKindOriginationNew:
		return prepareOrigination(ctx, cfgCtx, operation, withStorageDiff)
	case modelTypes.OperationKindSrExecuteOutboxMessage:
		return prepareSrExecute(ctx, cfgCtx, operation)
	case modelTypes.OperationKindRegisterGlobalConstant:
		response, _, err := newOperationResponse(ctx, cfgCtx, operation)
		return response, err
	case modelTypes.OperationKindSrOrigination:
		response, _, err := newOperationResponse(ctx, cfgCtx, operation)
		return response, err
	default:
		return Operation{}, errors.Errorf("unknown operation kind: %s", operation.Kind.String())
	}
}

// PrepareOperations -
func PrepareOperations(c context.Context, ctx *config.Context, ops []operation.Operation, withStorageDiff bool) ([]Operation, error) {
	resp := make([]Operation, len(ops))
	for i := 0; i < len(ops); i++ {
		op, err := prepareOperation(c, ctx, ops[i], withStorageDiff)
		if err != nil {
			return nil, err
		}
		resp[i] = op
	}
	return resp, nil
}

func setParameters(data []byte, parameter *ast.TypedAst, op *Operation) error {
	if len(data) == 0 {
		return nil
	}
	params := types.NewParameters(data)
	return setParatemetersWithType(params, parameter, op)
}

func setParatemetersWithType(params *types.Parameters, parameter *ast.TypedAst, op *Operation) error {
	if params == nil {
		return errors.New("Empty parameters")
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

func setStorageDiff(c context.Context, ctx *config.Context, destinationID int64, storage []byte, op *Operation, bmd []bigmapdiff.BigMapDiff, storageType *ast.TypedAst) error {
	storageDiff, err := getStorageDiff(c, ctx, destinationID, bmd, storage, storageType, op)
	if err != nil {
		return err
	}
	op.StorageDiff = storageDiff
	return nil
}

func getStorageDiff(c context.Context, ctx *config.Context, destinationID int64, bmd []bigmapdiff.BigMapDiff, storage []byte, storageType *ast.TypedAst, op *Operation) (*ast.MiguelNode, error) {
	currentStorage := &ast.TypedAst{
		Nodes: []ast.Node{ast.Copy(storageType.Nodes[0])},
	}
	var prevStorage *ast.TypedAst

	prev, err := ctx.Operations.Last(
		c,
		map[string]interface{}{
			"destination_id": destinationID,
			"status":         modelTypes.OperationStatusApplied,
			"timestamp": core.TimestampFilter{
				Lt: op.Timestamp,
			},
		}, op.ID)
	if err == nil {
		prevStorage = &ast.TypedAst{
			Nodes: []ast.Node{ast.Copy(storageType.Nodes[0])},
		}

		prevBmd, err := ctx.BigMapDiffs.Previous(c, bmd)
		if err != nil {
			return nil, err
		}

		if len(prev.DeffatedStorage) > 0 {
			if len(prevBmd) > 0 {
				if err := prepareStorage(prevStorage, prev.DeffatedStorage, prevBmd); err != nil {
					return nil, err
				}
			} else {
				if err := prepareStorage(prevStorage, prev.DeffatedStorage, nil); err != nil {
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
	if err := storageType.SettleFromBytes(deffatedStorage); err != nil {
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

func getErrorLocation(c context.Context, ctx *config.Context, operation operation.Operation, window int) (GetErrorLocationResponse, error) {
	proto, err := ctx.Cache.ProtocolByID(c, operation.ProtocolID)
	if err != nil {
		return GetErrorLocationResponse{}, err
	}
	code, err := getScriptBytes(c, ctx.Cache, operation.Destination.Address, proto.SymLink)
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
	start := helpers.Max(0, row-window)
	end := helpers.Min(len(rows), row+window+1)

	rows = rows[start:end]
	return GetErrorLocationResponse{
		Text:        strings.Join(rows, "\n"),
		FailedRow:   row + 1,
		StartColumn: sCol,
		EndColumn:   eCol,
		FirstRow:    start + 1,
	}, nil
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
