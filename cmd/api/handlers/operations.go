package handlers

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	formattererror "github.com/baking-bad/bcdhub/internal/contractparser/formatter_error"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/schema"
	"github.com/baking-bad/bcdhub/internal/tzkt"
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
// @Router /contract/{network}/{address}/operations [get]
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
	ops, err := ctx.Operations.GetByContract(req.Network, req.Address, filtersReq.Size, filters)
	if handleError(c, err, 0) {
		return
	}

	resp, err := PrepareOperations(ctx.BigMapDiffs, ctx.Schema, ctx.Operations, ops.Operations, filtersReq.WithStorageDiff)
	if handleError(c, err, 0) {
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
// @Accept  json
// @Produce  json
// @Success 200 {array} Operation
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /opg/{hash} [get]
func (ctx *Context) GetOperation(c *gin.Context) {
	var req OPGRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var queryReq opgRequest
	if err := c.BindQuery(&queryReq); handleError(c, err, http.StatusBadRequest) {
		return
	}

	op, err := ctx.Operations.Get(
		map[string]interface{}{
			"hash": req.Hash,
		},
		0,
		true,
	)
	if !core.IsRecordNotFound(err) && handleError(c, err, 0) {
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

	resp, err := PrepareOperations(ctx.BigMapDiffs, ctx.Schema, ctx.Operations, op, true)
	if handleError(c, err, 0) {
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
// @Router /operation/{id}/error_location [get]
func (ctx *Context) GetOperationErrorLocation(c *gin.Context) {
	var req getOperationByIDRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}
	operation := operation.Operation{ID: req.ID}
	if err := ctx.Storage.GetByID(&operation); handleError(c, err, 0) {
		return
	}

	if !cerrors.HasScriptRejectedError(operation.Errors) {
		handleError(c, errors.Errorf("No reject script error in operation"), http.StatusBadRequest)
		return
	}

	response, err := ctx.getErrorLocation(operation, 2)
	if handleError(c, err, 0) {
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

	operation := ctx.prepareMempoolOperation(res[0], network)
	ops <- operation
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

func formatErrors(errs []*cerrors.Error, op *Operation) error {
	for i := range errs {
		if err := errs[i].Format(); err != nil {
			return err
		}
	}
	op.Errors = errs
	return nil
}

func prepareOperation(bmdRepo bigmapdiff.Repository, schemaRepo schema.Repository, operationsRepo operation.Repository, operation operation.Operation, bmd []bigmapdiff.BigMapDiff, withStorageDiff bool) (Operation, error) {
	var op Operation
	op.FromModel(operation)

	var result OperationResult
	result.FromModel(operation.Result)
	op.Result = &result

	if err := formatErrors(operation.Errors, &op); err != nil {
		return op, err
	}
	if withStorageDiff {
		if operation.DeffatedStorage != "" && strings.HasPrefix(op.Destination, "KT") && op.Status == consts.Applied {
			if err := setStorageDiff(bmdRepo, schemaRepo, operationsRepo, op.Destination, operation.DeffatedStorage, &op, bmd); err != nil {
				return op, err
			}
		}
	}

	if op.Kind != consts.Transaction {
		return op, nil
	}

	if strings.HasPrefix(op.Destination, "KT") && !cerrors.HasParametersError(op.Errors) {
		if err := setParameters(schemaRepo, operation.Parameters, &op); err != nil {
			return op, err
		}
	}

	return op, nil
}

// PrepareOperations -
func PrepareOperations(bmdRepo bigmapdiff.Repository, schemaRepo schema.Repository, operationsRepo operation.Repository, ops []operation.Operation, withStorageDiff bool) ([]Operation, error) {
	resp := make([]Operation, len(ops))
	for i := 0; i < len(ops); i++ {
		var bmd []bigmapdiff.BigMapDiff
		var err error

		if withStorageDiff {
			bmd, err = bmdRepo.GetBigMapDiffsUniqueByOperationID(ops[i].ID)
			if err != nil {
				return nil, err
			}
		}

		op, err := prepareOperation(bmdRepo, schemaRepo, operationsRepo, ops[i], bmd, withStorageDiff)
		if err != nil {
			return nil, err
		}
		resp[i] = op
	}
	return resp, nil
}

func setParameters(schemaRepo schema.Repository, parameters string, op *Operation) error {
	metadata, err := meta.GetMetadata(schemaRepo, op.Destination, consts.PARAMETER, op.Protocol)
	if err != nil {
		return nil
	}

	params := gjson.Parse(parameters)
	op.Parameters, err = newmiguel.ParameterToMiguel(params, metadata)
	if err != nil {
		if !cerrors.HasGasExhaustedError(op.Errors) {
			helpers.CatchErrorSentry(err)
			return err
		}
	}
	return nil
}

func setStorageDiff(bmdRepo bigmapdiff.Repository, schemaRepo schema.Repository, operationsRepo operation.Repository, address, storage string, op *Operation, bmd []bigmapdiff.BigMapDiff) error {
	metadata, err := meta.GetContractMetadata(schemaRepo, address)
	if err != nil {
		return err
	}
	storageDiff, err := getStorageDiff(bmdRepo, operationsRepo, bmd, address, storage, metadata, false, op)
	if err != nil {
		return err
	}
	op.StorageDiff = storageDiff
	return nil
}

func getStorageDiff(bmdRepo bigmapdiff.Repository, operationsRepo operation.Repository, bmd []bigmapdiff.BigMapDiff, address, storage string, metadata *meta.ContractMetadata, isSimulating bool, op *Operation) (interface{}, error) {
	var prevStorage *newmiguel.Node
	var prevDeffatedStorage string
	prev, err := operationsRepo.Last(address, op.Network, op.IndexedTime)
	if err == nil {
		prevBmd, err := getPrevBmd(bmdRepo, bmd, op.IndexedTime, op.Destination)
		if err != nil {
			return nil, err
		}

		var exDeffatedStorage string
		exOp, err := operationsRepo.Last(address, prev.Network, prev.IndexedTime)
		if err == nil {
			exDeffatedStorage = exOp.DeffatedStorage
		} else if !core.IsRecordNotFound(err) {
			return nil, err
		}

		prevStorage, err = getEnrichStorageMiguel(prevBmd, prev.Protocol, prev.DeffatedStorage, exDeffatedStorage, metadata, isSimulating)
		if err != nil {
			return nil, err
		}
		prevDeffatedStorage = prev.DeffatedStorage
	} else {
		if !core.IsRecordNotFound(err) {
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
	return currentStorage, nil
}

func getEnrichStorageMiguel(bmd []bigmapdiff.BigMapDiff, protocol, storage, prevStorage string, metadata *meta.ContractMetadata, isSimulating bool) (*newmiguel.Node, error) {
	store, err := enrichStorage(storage, prevStorage, bmd, protocol, false, isSimulating)
	if err != nil {
		return nil, err
	}
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

	parser, err := contractparser.MakeStorageParser(nil, nil, protocol, isSimulating)
	if err != nil {
		return gjson.Result{}, err
	}

	return parser.Enrich(s, prevStorage, bmd, skipEmpty, true)
}

func getPrevBmd(repo bigmapdiff.Repository, bmd []bigmapdiff.BigMapDiff, indexedTime int64, address string) ([]bigmapdiff.BigMapDiff, error) {
	if len(bmd) == 0 {
		return nil, nil
	}
	return repo.GetBigMapDiffsPrevious(bmd, indexedTime, address)
}

func (ctx *Context) prepareMempoolOperation(item tzkt.MempoolOperation, network string) *Operation {
	status := item.Body.Status
	if status == consts.Applied {
		status = "pending"
	}

	if !helpers.StringInArray(item.Body.Kind, []string{consts.Transaction, consts.Origination, consts.OriginationNew}) {
		return nil
	}

	op := Operation{
		Protocol:  item.Body.Protocol,
		Hash:      item.Body.Hash,
		Network:   network,
		Timestamp: time.Unix(item.Body.Timestamp, 0).UTC(),

		Kind:         item.Body.Kind,
		Source:       item.Body.Source,
		Fee:          item.Body.Fee,
		Counter:      item.Body.Counter,
		GasLimit:     item.Body.GasLimit,
		StorageLimit: item.Body.StorageLimit,
		Amount:       item.Body.Amount,
		Destination:  item.Body.Destination,
		Mempool:      true,
		Status:       status,
		RawMempool:   string(item.Raw),
	}

	aliases, err := ctx.ES.GetAliasesMap(network)
	if err != nil {
		if !elastic.IsRecordNotFound(err) {
			return &op
		}
	} else {
		op.SourceAlias = aliases[op.Source]
		op.DestinationAlias = aliases[op.Destination]
	}
	errs, err := cerrors.ParseArray(item.Body.Errors)
	if err != nil {
		return nil
	}
	op.Errors = errs

	if op.Kind != consts.Transaction {
		return &op
	}

	if helpers.IsContract(op.Destination) && op.Protocol != "" {
		if params := gjson.ParseBytes(item.Body.Parameters); params.Exists() {
			ctx.buildOperationParameters(params, &op)
		} else {
			op.Entrypoint = consts.DefaultEntrypoint
		}
	}

	return &op
}

func (ctx *Context) buildOperationParameters(params gjson.Result, op *Operation) {
	metadata, err := meta.GetMetadata(ctx.Schema, op.Destination, consts.PARAMETER, op.Protocol)
	if err != nil {
		return
	}

	op.Entrypoint, err = metadata.GetByPath(params)
	if err != nil && op.Errors == nil {
		return
	}

	op.Parameters, err = newmiguel.ParameterToMiguel(params, metadata)
	if err != nil {
		if !cerrors.HasParametersError(op.Errors) {
			return
		}
	}
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
	opErr := cerrors.First(operation.Errors, consts.ScriptRejectedError)
	if opErr == nil {
		return GetErrorLocationResponse{}, errors.Errorf("Can't find script rejected error")
	}
	defaultError, ok := opErr.IError.(*cerrors.DefaultError)
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
