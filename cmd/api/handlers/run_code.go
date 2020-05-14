package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
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

	log.Println(response)

	main := models.Operation{
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

	// if err := setParameters(ctx.ES, input.Raw, &main); handleError(c, err, 0) {
	// 	return
	// }
	operations, err := ctx.parseRunCodeResponse(response, &main)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, operations)
}

func (ctx *Context) parseRunCodeResponse(response gjson.Result, main *models.Operation) ([]models.Operation, error) {
	if response.IsArray() {
		return ctx.parseFailedRunCode(response, main)
	} else if response.IsObject() {
		return ctx.parseAppliedRunCode(response, main)
	}
	return nil, fmt.Errorf("Unknown response: %v", response.Value())
}

func (ctx *Context) parseFailedRunCode(response gjson.Result, main *models.Operation) ([]models.Operation, error) {
	operations := make([]models.Operation, 0)
	// main.Errors = cerrors.ParseArray(response)
	// if err := formatErrors(main.Errors, main); err != nil {
	// 	return nil, err
	// }
	// main.Status = "failed"
	// operations = append(operations, *main)
	return operations, nil
}

func (ctx *Context) parseAppliedRunCode(response gjson.Result, main *models.Operation) ([]models.Operation, error) {
	operations := make([]models.Operation, 0)

	// metadata, err := meta.GetContractMetadata(ctx.ES, main.Destination)
	// if err != nil {
	// 	return nil, err
	// }
	// bmd, err := ctx.parseBigMapDiffs(response, main)
	// if err != nil {
	// 	return err
	// }

	// operations = append(operations, *main)

	// operationsJSON := response.Get("operations").Array()
	// for _, item := range operationsJSON {
	// 	var op models.Operation
	// 	op.ParseElasticJSON(item)
	// 	op.Status = "applied"
	// 	op.Network = main.Network
	// 	op.Timestamp = main.Timestamp
	// 	op.Protocol = main.Protocol
	// 	op.Level = main.Level
	// 	op.Internal = true
	// 	operations = append(operations, op)
	// }
	return operations, nil
}

// func (ctx *Context) parseBigMapDiffs(response gjson.Result, metadata meta.Metadata, operation *Operation) ([]models.BigMapDiff, error) {
// 	rpc, err := ctx.GetRPC(operation.Network)
// 	if err != nil {
// 		return nil, err
// 	}
// 	parser := storage.NewSimulate(rpc, ctx.ES)
// 	switch operation.Kind {
// 	case consts.Transaction:
// 		parser.ParseTransaction(response, metadata)
// 	case consts.Origination:
// 	}
// }
