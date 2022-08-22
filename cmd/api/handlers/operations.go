package handlers

import (
	"net/http"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	formattererror "github.com/baking-bad/bcdhub/internal/bcd/formatter/error"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/config"
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
func GetContractOperations() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getAccountRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		var filtersReq operationsRequest
		if err := c.BindQuery(&filtersReq); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		account, err := ctx.Accounts.Get(req.Address)
		if handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		filters := prepareFilters(filtersReq)
		ops, err := ctx.Operations.GetByAccount(account, filtersReq.Size, filters)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		resp, err := PrepareOperations(ctx, ops.Operations, filtersReq.WithStorageDiff)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		c.SecureJSON(http.StatusOK, OperationResponse{
			Operations: resp,
			LastID:     ops.LastID,
		})
	}
}

// GetOperation godoc
// @Summary Get operation group
// @Description Get operation group by hash
// @Tags operations
// @ID get-opg
// @Param hash path string true "Operation group hash"  minlength(51) maxlength(51)
// @Param with_mempool query bool false "Search operation in mempool or not"
// @Param with_storage_diff query bool false "Include storage diff to operations or not"
// @Param network query string false "Network"
// @Accept  json
// @Produce  json
// @Success 200 {array} Operation
// @Success 204 {object} gin.H
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/opg/{hash} [get]
func GetOperation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctxs := c.MustGet("contexts").(config.Contexts)
		any := ctxs.Any()

		var req OPGRequest
		if err := c.BindUri(&req); handleError(c, any.Storage, err, http.StatusBadRequest) {
			return
		}

		var queryReq opgRequest
		if err := c.BindQuery(&queryReq); handleError(c, any.Storage, err, http.StatusBadRequest) {
			return
		}

		operations := make([]operation.Operation, 0)
		var foundContext *config.Context

		network := modelTypes.NewNetwork(queryReq.Network)
		if ctx, ok := ctxs[network]; ok {
			op, err := ctx.Operations.GetByHash(req.Hash)
			if err != nil {
				if !ctx.Storage.IsRecordNotFound(err) {
					handleError(c, ctx.Storage, err, 0)
					return
				}
			} else {
				foundContext = ctx
				operations = append(operations, op...)
			}
		} else {
			for _, ctx := range ctxs {
				op, err := ctx.Operations.GetByHash(req.Hash)
				if err != nil {
					if !ctx.Storage.IsRecordNotFound(err) {
						handleError(c, ctx.Storage, err, 0)
						return
					}
					continue
				}
				operations = append(operations, op...)
				if len(operations) > 0 {
					foundContext = ctx
					break
				}
			}
		}

		if foundContext == nil {
			opg := make([]Operation, 0)

			if queryReq.WithMempool {
				ctx := ctxs.Any()
				operation, err := getOperationFromMempool(ctx, req.Hash)
				if handleError(c, ctx.Storage, err, 0) {
					return
				}
				if operation != nil {
					opg = append(opg, *operation)
				}
			}

			if len(opg) == 0 {
				c.SecureJSON(http.StatusNoContent, []gin.H{})
				return
			}

			c.SecureJSON(http.StatusOK, opg)
			return
		}

		resp, err := PrepareOperations(foundContext, operations, queryReq.WithStorageDiff)
		if handleError(c, foundContext.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, resp)
	}
}

// GetImplicitOperation godoc
// @Summary Get implicit operation
// @DescriptionGet implicit operation
// @Tags operations
// @ID get-implicit-operation
// @Param network path string true "Network"
// @Param counter path integer true "Counter"
// @Accept  json
// @Produce  json
// @Success 200 {array} Operation
// @Success 204 {object} gin.H
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/implicit/{network}/{counter} [get]
func GetImplicitOperation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req ImplicitOperationRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		op, err := ctx.Operations.GetImplicitOperation(req.Counter)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		resp, err := PrepareOperations(ctx, []operation.Operation{op}, false)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, resp)
	}
}

// GetOperationErrorLocation godoc
// @Summary Get code line where operation failed
// @Description Get code line where operation failed
// @Tags operations
// @ID get-operation-error-location
// @Param network path string true "Network"
// @Param id path integer true "Internal BCD operation ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} GetErrorLocationResponse
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/operation/{network}/{id}/error_location [get]
func GetOperationErrorLocation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getOperationByIDRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		operation, err := ctx.Operations.GetByID(req.ID)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		if !tezerrors.HasScriptRejectedError(operation.Errors) {
			handleError(c, ctx.Storage, errors.Errorf("No reject script error in operation"), http.StatusBadRequest)
			return
		}

		response, err := getErrorLocation(ctx, operation, 2)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		c.SecureJSON(http.StatusOK, response)
	}
}

// GetOperationDiff godoc
// @Summary Get operation storage diff
// @DescriptionGet Get operation storage diff
// @Tags operations
// @ID get-operation-diff
// @Param network path string true "Network"
// @Param id path integer true "Internal BCD operation ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} ast.MiguelNode
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/operation/{network}/{id}/diff [get]
func GetOperationDiff() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getOperationByIDRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}
		operation, err := ctx.Operations.GetByID(req.ID)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		var result Operation
		result.FromModel(operation)

		if len(operation.DeffatedStorage) > 0 && (operation.IsCall() || operation.IsOrigination() || operation.IsImplicit()) && operation.IsApplied() {
			proto, err := ctx.Cache.ProtocolByID(operation.ProtocolID)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			result.Protocol = proto.Hash

			storageBytes, err := ctx.Contracts.ScriptPart(operation.Destination.Address, proto.SymLink, consts.STORAGE)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}

			storageType, err := ast.NewTypedAstFromBytes(storageBytes)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}

			bmd, err := ctx.BigMapDiffs.GetForOperation(operation.ID)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}

			if err := setStorageDiff(ctx, operation.DestinationID, operation.DeffatedStorage, &result, bmd, storageType); handleError(c, ctx.Storage, err, 0) {
				return
			}
		}
		c.SecureJSON(http.StatusOK, result.StorageDiff)
	}
}

// GetOperationGroups -
// @Summary Get operation groups by account
// @Description Get operation groups by account
// @Tags contract
// @ID get-operation-groups-by-account
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param last_id query string false "Last operation ID"
// @Param size query integer false "Expected OPG count" mininum(1)
// @Accept  json
// @Produce  json
// @Success 200 {array} OPGResponse
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/opg [get]
func GetOperationGroups() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getAccountRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		var args opgForAddressRequest
		if err := c.BindQuery(&args); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		opg, err := ctx.Operations.OPG(req.Address, int64(args.Size), args.LastID)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		response := make([]OPGResponse, len(opg))
		for i := range opg {
			response[i] = NewOPGResponse(opg[i])
		}
		c.SecureJSON(http.StatusOK, response)
	}
}

// GetByHashAndCounter -
// @Summary Get operations by hash and counter
// @Description Get operations by hash and counter
// @Tags operations
// @ID get-operations-by-hash-and-counter
// @Param hash path string true "Operation group hash"  minlength(51) maxlength(51)
// @Param counter path integer true "Counter of main operation"
// @Param network query string false "You can set network field for better performance"
// @Accept  json
// @Produce  json
// @Success 200 {array} Operation
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/opg/{hash}/{counter} [get]
func GetByHashAndCounter() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctxs := c.MustGet("contexts").(config.Contexts)

		var req OperationGroupContentRequest
		if err := c.BindUri(&req); handleError(c, ctxs.Any().Storage, err, http.StatusNotFound) {
			return
		}

		var args networkQueryRequest
		if err := c.BindQuery(&args); handleError(c, ctxs.Any().Storage, err, http.StatusBadRequest) {
			return
		}

		var opg []operation.Operation
		var foundContext *config.Context

		ctx, err := ctxs.Get(modelTypes.NewNetwork(args.Network))
		if err == nil {
			opg, err = ctx.Operations.GetByHashAndCounter(req.Hash, req.Counter)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			foundContext = ctx
		} else {
			for _, ctx := range ctxs {
				opg, err = ctx.Operations.GetByHashAndCounter(req.Hash, req.Counter)
				if handleError(c, ctx.Storage, err, 0) {
					return
				}
				if len(opg) > 0 {
					foundContext = ctx
					break
				}
			}
		}

		resp, err := PrepareOperations(foundContext, opg, false)
		if handleError(c, foundContext.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, resp)
	}
}

func getOperationFromMempool(ctx *config.Context, hash string) (*Operation, error) {
	res, err := ctx.Mempool.GetByHash(hash)
	if err != nil {
		return nil, err
	}

	switch {
	case len(res.Originations) > 0:
		return prepareMempoolOrigination(ctx, res.Originations[0]), nil
	case len(res.Transactions) > 0:
		return prepareMempoolTransaction(ctx, res.Transactions[0]), nil
	default:
		return nil, nil
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

func prepareOperation(ctx *config.Context, operation operation.Operation, bmd []bigmapdiff.BigMapDiff, withStorageDiff bool) (Operation, error) {
	var op Operation
	op.FromModel(operation)

	op.SourceAlias = operation.Source.Alias
	op.DestinationAlias = operation.Destination.Alias

	proto, err := ctx.Cache.ProtocolByID(operation.ProtocolID)
	if err != nil {
		return op, err
	}
	op.Protocol = proto.Hash

	if operation.IsEvent() {
		eventType, err := ast.NewTypedAstFromBytes(operation.EventType)
		if err != nil {
			return op, err
		}
		if err := eventType.SettleFromBytes(operation.EventPayload); err != nil {
			return op, err
		}
		eventMiguel, err := eventType.ToMiguel()
		if err != nil {
			return op, err
		}
		op.Event = eventMiguel
		return op, err
	}

	if bcd.IsContract(op.Destination) {
		if err := formatErrors(operation.Errors, &op); err != nil {
			return op, err
		}

		if withStorageDiff {
			storageType, err := getStorageType(ctx.Contracts, op.Destination, proto.SymLink)
			if err != nil {
				return op, err
			}
			if len(operation.DeffatedStorage) > 0 && (operation.IsCall() || operation.IsOrigination() || operation.IsImplicit()) && operation.IsApplied() {
				if err := setStorageDiff(ctx, operation.DestinationID, operation.DeffatedStorage, &op, bmd, storageType); err != nil {
					return op, err
				}
			}
		}

		if !operation.IsTransaction() {
			return op, nil
		}

		if operation.IsCall() && !tezerrors.HasParametersError(op.Errors) {
			parameterType, err := getParameterType(ctx.Contracts, op.Destination, proto.SymLink)
			if err != nil {
				return op, err
			}
			if err := setParameters(operation.Parameters, parameterType, &op); err != nil {
				return op, err
			}
		}
	}

	return op, nil
}

// PrepareOperations -
func PrepareOperations(ctx *config.Context, ops []operation.Operation, withStorageDiff bool) ([]Operation, error) {
	resp := make([]Operation, len(ops))
	for i := 0; i < len(ops); i++ {
		var diffs []bigmapdiff.BigMapDiff
		var err error

		if withStorageDiff {
			diffs, err = ctx.BigMapDiffs.GetForOperation(ops[i].ID)
			if err != nil {
				return nil, err
			}
		}

		op, err := prepareOperation(ctx, ops[i], diffs, withStorageDiff)
		if err != nil {
			return nil, err
		}
		op.Network = ctx.Network.String()
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

func setStorageDiff(ctx *config.Context, destinationID int64, storage []byte, op *Operation, bmd []bigmapdiff.BigMapDiff, storageType *ast.TypedAst) error {
	storageDiff, err := getStorageDiff(ctx, destinationID, bmd, storage, storageType, op)
	if err != nil {
		return err
	}
	op.StorageDiff = storageDiff
	return nil
}

func getStorageDiff(ctx *config.Context, destinationID int64, bmd []bigmapdiff.BigMapDiff, storage []byte, storageType *ast.TypedAst, op *Operation) (*ast.MiguelNode, error) {
	currentStorage := &ast.TypedAst{
		Nodes: []ast.Node{ast.Copy(storageType.Nodes[0])},
	}
	var prevStorage *ast.TypedAst

	prev, err := ctx.Operations.Last(
		map[string]interface{}{
			"destination_id": destinationID,
			"status":         modelTypes.OperationStatusApplied,
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

func getErrorLocation(ctx *config.Context, operation operation.Operation, window int) (GetErrorLocationResponse, error) {
	proto, err := ctx.Cache.ProtocolByID(operation.ProtocolID)
	if err != nil {
		return GetErrorLocationResponse{}, err
	}
	code, err := getScriptBytes(ctx.Contracts, operation.Destination.Address, proto.SymLink)
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
