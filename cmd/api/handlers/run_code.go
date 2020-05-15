package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// RunCode -
func (ctx *Context) RunCode(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}
	var reqRunCode runCodeRequest
	if err := c.BindJSON(&reqRunCode); handleError(c, err, http.StatusBadRequest) {
		return
	}

	rpc, err := ctx.GetRPC(req.Network)
	if handleError(c, err, http.StatusBadRequest) {
		return
	}

	state, err := ctx.ES.CurrentState(req.Network)
	if handleError(c, err, 0) {
		return
	}

	script, err := contractparser.GetContract(rpc, req.Address, req.Network, state.Protocol, ctx.SharePath, 0)
	if handleError(c, err, 0) {
		return
	}

	input, err := ctx.buildEntrypointMicheline(req.Network, req.Address, reqRunCode.BinPath, reqRunCode.Data)
	if handleError(c, err, 0) {
		return
	}

	if !input.Get("entrypoint").Exists() || !input.Get("value").Exists() {
		handleError(c, fmt.Errorf("Error during build parameters: %s", input.String()), 0)
		return
	}

	entrypoint := input.Get("entrypoint").String()
	value := input.Get("value")

	storage, err := rpc.GetScriptStorageJSON(req.Address, 0)
	if handleError(c, err, 0) {
		return
	}

	response, err := rpc.RunCode(script.Get("code"), storage, value, state.ChainID, reqRunCode.Source, reqRunCode.Sender, entrypoint, reqRunCode.Amount, reqRunCode.GasLimit)
	if handleError(c, err, 0) {
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
		Status:      "applied",
		Entrypoint:  entrypoint,
	}

	if err := setParameters(ctx.ES, input.Raw, &main); handleError(c, err, 0) {
		return
	}
	if err := ctx.setSimulateStorageDiff(response, &main); handleError(c, err, 0) {
		return
	}
	operations, err := ctx.parseRunCodeResponse(response, &main)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, operations)
}

func (ctx *Context) parseRunCodeResponse(response gjson.Result, main *Operation) ([]Operation, error) {
	if response.IsArray() {
		return ctx.parseFailedRunCode(response, main)
	} else if response.IsObject() {
		return ctx.parseAppliedRunCode(response, main)
	}
	return nil, fmt.Errorf("Unknown response: %v", response.Value())
}

func (ctx *Context) parseFailedRunCode(response gjson.Result, main *Operation) ([]Operation, error) {
	main.Errors = cerrors.ParseArray(response)
	if err := formatErrors(main.Errors, main); err != nil {
		return nil, err
	}
	main.Status = "failed"
	return []Operation{*main}, nil
}

func (ctx *Context) parseAppliedRunCode(response gjson.Result, main *Operation) ([]Operation, error) {
	operations := []Operation{*main}

	operationsJSON := response.Get("operations").Array()
	for _, item := range operationsJSON {
		var op Operation
		op.Kind = item.Get("kind").String()
		op.Amount = item.Get("amount").Int()
		op.Source = item.Get("source").String()
		op.Destination = item.Get("destination").String()
		op.Status = "applied"
		op.Network = main.Network
		op.Timestamp = main.Timestamp
		op.Protocol = main.Protocol
		op.Level = main.Level
		op.Internal = true
		if err := setParameters(ctx.ES, item.Get("parameters").Raw, &op); err != nil {
			return nil, err
		}
		if err := ctx.setSimulateStorageDiff(item, &op); err != nil {
			return nil, err
		}
		operations = append(operations, op)
	}
	return operations, nil
}

func (ctx *Context) parseBigMapDiffs(response gjson.Result, metadata meta.Metadata, operation *Operation) ([]models.BigMapDiff, error) {
	rpc, err := ctx.GetRPC(operation.Network)
	if err != nil {
		return nil, err
	}

	model := operation.ToModel()
	parser := storage.NewSimulate(rpc, ctx.ES)

	rs := storage.RichStorage{Empty: true}
	switch operation.Kind {
	case consts.Transaction:
		rs, err = parser.ParseTransaction(response, metadata, model)
	case consts.Origination:
		rs, err = parser.ParseOrigination(response, metadata, model)
	}
	if err != nil {
		return nil, err
	}
	if rs.Empty {
		return nil, nil
	}
	bmd := make([]models.BigMapDiff, len(rs.BigMapDiffs))
	for i := range rs.BigMapDiffs {
		bmd[i] = *rs.BigMapDiffs[i]
	}
	return bmd, nil
}

func (ctx *Context) setSimulateStorageDiff(response gjson.Result, main *Operation) error {
	storage := response.Get("storage").String()
	if storage == "" || !strings.HasPrefix(main.Destination, "KT") || main.Status != "applied" {
		return nil
	}
	metadata, err := meta.GetContractMetadata(ctx.ES, main.Destination)
	if err != nil {
		return err
	}
	storageMetadata, err := metadata.Get(consts.STORAGE, main.Protocol)
	if err != nil {
		return err
	}
	bmd, err := ctx.parseBigMapDiffs(response, storageMetadata, main)
	if err != nil {
		return err
	}
	storageDiff, err := getStorageDiff(ctx.ES, bmd, main.Destination, storage, metadata, true, main)
	if err != nil {
		return err
	}
	main.StorageDiff = storageDiff
	return nil
}
