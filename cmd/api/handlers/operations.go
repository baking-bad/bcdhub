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
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
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
	ops, err := ctx.ES.GetOperationsForContract(req.Network, req.Address, filtersReq.Size, filters)
	if handleError(c, err, 0) {
		return
	}

	resp, err := PrepareOperations(ctx.ES, ops.Operations)
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

	op, err := ctx.ES.GetOperations(
		map[string]interface{}{
			"hash": req.Hash,
		},
		0,
		true,
	)
	if !elastic.IsRecordNotFound(err) && handleError(c, err, 0) {
		return
	}

	if len(op) == 0 {
		operation, err := ctx.getOperationFromMempool(req.Hash)
		if handleError(c, err, http.StatusNotFound) {
			return
		}

		c.JSON(http.StatusOK, []Operation{operation})
		return
	}

	resp, err := PrepareOperations(ctx.ES, op)
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
	operation := models.Operation{ID: req.ID}
	if err := ctx.ES.GetByID(&operation); handleError(c, err, 0) {
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

func (ctx *Context) getOperationFromMempool(hash string) (Operation, error) {
	var wg sync.WaitGroup
	var opCh = make(chan Operation, len(ctx.TzKTServices))

	defer close(opCh)

	for network := range ctx.TzKTServices {
		wg.Add(1)
		go ctx.getOperation(network, hash, opCh, &wg)
	}

	wg.Wait()

	return <-opCh, nil
}

func (ctx *Context) getOperation(network, hash string, ops chan<- Operation, wg *sync.WaitGroup) {
	defer wg.Done()

	api, err := ctx.GetTzKTService(network)
	if err != nil {
		return
	}

	res, err := api.GetMempool(hash)
	if err != nil {
		return
	}

	if res.Get("#").Int() == 0 {
		return
	}

	operation, err := ctx.prepareMempoolOperation(res, network, hash)
	if err != nil {
		return
	}

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

func formatErrors(errs []cerrors.IError, op *Operation) error {
	for i := range errs {
		if err := errs[i].Format(); err != nil {
			return err
		}
	}
	op.Errors = errs
	return nil
}

func prepareOperation(es elastic.IElastic, operation models.Operation, bmd []models.BigMapDiff) (Operation, error) {
	var op Operation
	op.FromModel(operation)

	var result OperationResult
	result.FromModel(operation.Result)
	op.Result = &result

	if err := formatErrors(operation.Errors, &op); err != nil {
		return op, err
	}
	if operation.DeffatedStorage != "" && strings.HasPrefix(op.Destination, "KT") && op.Status == "applied" {
		if err := setStorageDiff(es, op.Destination, op.Network, operation.DeffatedStorage, &op, bmd); err != nil {
			return op, err
		}
	}

	if op.Kind != consts.Transaction {
		return op, nil
	}

	if strings.HasPrefix(op.Destination, "KT") && !cerrors.HasParametersError(op.Errors) {
		if err := setParameters(es, operation.Parameters, &op); err != nil {
			return op, err
		}
	}

	return op, nil
}

// PrepareOperations -
func PrepareOperations(es elastic.IElastic, ops []models.Operation) ([]Operation, error) {
	resp := make([]Operation, len(ops))
	for i := 0; i < len(ops); i++ {
		bmd, err := es.GetBigMapDiffsUniqueByOperationID(ops[i].ID)
		if err != nil {
			return nil, err
		}
		op, err := prepareOperation(es, ops[i], bmd)
		if err != nil {
			return nil, err
		}
		resp[i] = op
	}
	return resp, nil
}

func setParameters(es elastic.IElastic, parameters string, op *Operation) error {
	metadata, err := meta.GetMetadata(es, op.Destination, consts.PARAMETER, op.Protocol)
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

func setStorageDiff(es elastic.IElastic, address, network, storage string, op *Operation, bmd []models.BigMapDiff) error {
	metadata, err := meta.GetContractMetadata(es, address)
	if err != nil {
		return err
	}
	storageDiff, err := getStorageDiff(es, bmd, address, storage, metadata, false, op)
	if err != nil {
		return err
	}
	op.StorageDiff = storageDiff
	return nil
}

func getStorageDiff(es elastic.IElastic, bmd []models.BigMapDiff, address, storage string, metadata *meta.ContractMetadata, isSimulating bool, op *Operation) (interface{}, error) {
	var prevStorage *newmiguel.Node
	var prevDeffatedStorage string
	prev, err := es.GetLastOperation(address, op.Network, op.IndexedTime)
	if err == nil {
		prevBmd, err := getPrevBmd(es, bmd, op.IndexedTime, op.Destination)
		if err != nil {
			return nil, err
		}

		var exDeffatedStorage string
		exOp, err := es.GetLastOperation(address, prev.Network, prev.IndexedTime)
		if err == nil {
			exDeffatedStorage = exOp.DeffatedStorage
		} else {
			if !elastic.IsRecordNotFound(err) {
				return nil, err
			}
		}

		prevStorage, err = getEnrichStorageMiguel(es, prevBmd, prev.Protocol, prev.DeffatedStorage, exDeffatedStorage, metadata, isSimulating)
		if err != nil {
			return nil, err
		}
		prevDeffatedStorage = prev.DeffatedStorage
	} else {
		if !elastic.IsRecordNotFound(err) {
			return nil, err
		}
		prevStorage = nil
	}

	currentStorage, err := getEnrichStorageMiguel(es, bmd, op.Protocol, storage, prevDeffatedStorage, metadata, isSimulating)
	if err != nil {
		return nil, err
	}
	if currentStorage == nil {
		return nil, nil
	}

	currentStorage.Diff(prevStorage)
	return currentStorage, nil
}

func getEnrichStorageMiguel(es elastic.IElastic, bmd []models.BigMapDiff, protocol, storage, prevStorage string, metadata *meta.ContractMetadata, isSimulating bool) (*newmiguel.Node, error) {
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

func enrichStorage(s, prevStorage string, bmd []models.BigMapDiff, protocol string, skipEmpty, isSimulating bool) (gjson.Result, error) {
	if len(bmd) == 0 {
		return gjson.Parse(s), nil
	}

	parser, err := contractparser.MakeStorageParser(nil, nil, protocol, isSimulating)
	if err != nil {
		return gjson.Result{}, err
	}

	return parser.Enrich(s, prevStorage, bmd, skipEmpty)
}

func getPrevBmd(es elastic.IElastic, bmd []models.BigMapDiff, indexedTime int64, address string) ([]models.BigMapDiff, error) {
	if len(bmd) == 0 {
		return nil, nil
	}
	return es.GetBigMapDiffsPrevious(bmd, indexedTime, address)
}

func (ctx *Context) prepareMempoolOperation(res gjson.Result, network, hash string) (Operation, error) {
	item := res.Array()[0]

	status := item.Get("status").String()
	if status == "applied" {
		status = "pending"
	}

	op := Operation{
		Protocol:  item.Get("protocol").String(),
		Hash:      item.Get("hash").String(),
		Network:   network,
		Timestamp: time.Unix(item.Get("timestamp").Int(), 0).UTC(),

		Kind:         item.Get("kind").String(),
		Source:       item.Get("source").String(),
		Fee:          item.Get("fee").Int(),
		Counter:      item.Get("counter").Int(),
		GasLimit:     item.Get("gas_limit").Int(),
		StorageLimit: item.Get("storage_limit").Int(),
		Amount:       item.Get("amount").Int(),
		Destination:  item.Get("destination").String(),
		Mempool:      true,
		Status:       status,
		RawMempool:   item.Value(),
	}

	op.SourceAlias = ctx.Aliases[op.Source]
	op.DestinationAlias = ctx.Aliases[op.Destination]
	op.Errors = cerrors.ParseArray(item.Get("errors"))

	if op.Kind != consts.Transaction {
		return op, nil
	}

	if strings.HasPrefix(op.Destination, "KT") && op.Protocol != "" {
		if params := item.Get("parameters"); params.Exists() {
			ctx.buildOperationParameters(params, &op)
		} else {
			op.Entrypoint = "default"
		}
	}

	return op, nil
}

func (ctx *Context) buildOperationParameters(params gjson.Result, op *Operation) {
	metadata, err := meta.GetMetadata(ctx.ES, op.Destination, consts.PARAMETER, op.Protocol)
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

func (ctx *Context) getErrorLocation(operation models.Operation, window int) (GetErrorLocationResponse, error) {
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
		return GetErrorLocationResponse{}, errors.Errorf("Can't find script rejevted error")
	}
	defaultError, ok := opErr.(*cerrors.DefaultError)
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
