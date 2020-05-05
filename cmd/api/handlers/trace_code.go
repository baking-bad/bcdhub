package handlers

import (
	"fmt"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/gin-gonic/gin"
)

// TraceCode -
func (ctx *Context) TraceCode(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}
	var reqTraceCode traceCodeRequest
	if err := c.BindJSON(&reqTraceCode); handleError(c, err, http.StatusBadRequest) {
		return
	}

	rpc, ok := ctx.RPCs[req.Network]
	if !ok {
		handleError(c, fmt.Errorf("Unknown network: %s", req.Network), http.StatusBadRequest)
		return
	}
	state, err := ctx.ES.CurrentState(req.Network)
	if handleError(c, err, 0) {
		return
	}

	code, err := contractparser.GetContract(rpc, req.Address, req.Network, state.Protocol, ctx.Dir, 0)
	if handleError(c, err, 0) {
		return
	}

	input, err := ctx.buildEntrypointMicheline(req.Network, req.Address, reqTraceCode.BinPath, reqTraceCode.Data)
	if handleError(c, err, 0) {
		return
	}

	// TODO: storage with big map or not?
	response, err := rpc.TraceCode(code.String(), "", input.String(), state.ChainID, reqTraceCode.Source, reqTraceCode.Sender, "", reqTraceCode.Amount, reqTraceCode.GasLimit)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, response)
}
