package handlers

import (
	"fmt"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetContractCode -
func (ctx *Context) GetContractCode(c *gin.Context) {
	var req getContractCodeRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	if err := c.BindQuery(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	code, err := ctx.getContractCode(req.Network, req.Address, req.Protocol, req.Level)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, code)
}

// GetDiff -
func (ctx *Context) GetDiff(c *gin.Context) {
	var req getDiffRequest
	if err := c.BindQuery(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	d, err := ctx.getDiff(
		req.SourceAddress,
		req.SourceNetwork,
		req.DestinationAddress,
		req.DestinationNetwork,
		consts.CurrentProto,
		consts.CurrentProto,
	)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, d)
}

func (ctx *Context) getContractCode(network, address, protocol string, fallbackLevel int64) (string, error) {
	contract, err := ctx.getContractCodeJSON(network, address, protocol, fallbackLevel)
	if err != nil {
		return "", err
	}

	code := contract.Get("code")
	return formatter.MichelineToMichelson(code, false, formatter.DefLineSize)
}

func (ctx *Context) getContractCodeJSON(network, address, protocol string, fallbackLevel int64) (res gjson.Result, err error) {
	rpc, ok := ctx.RPCs[network]
	if !ok {
		return res, fmt.Errorf("Unknown network %s", network)
	}
	contract, err := contractparser.GetContract(rpc, address, network, protocol, ctx.Dir, fallbackLevel)
	if err != nil {
		return
	}
	if !contract.IsArray() && !contract.IsObject() {
		return res, fmt.Errorf("Unknown contract: %s", address)
	}

	// return macros.FindMacros(contractJSON)
	return contract, nil
}

func (ctx *Context) getDiff(srcAddress, srcNetwork, destAddress, destNetwork string, srcProtocol, destProtocol string) (res formatter.DiffResult, err error) {
	srcCode, err := ctx.getContractCodeJSON(srcNetwork, srcAddress, srcProtocol, 0) // fallbackLevel: head
	if err != nil {
		return
	}
	destCode, err := ctx.getContractCodeJSON(destNetwork, destAddress, destProtocol, 0) // fallbackLevel: head
	if err != nil {
		return
	}
	a := srcCode.Get("code")
	b := destCode.Get("code")
	res, err = formatter.Diff(a, b)
	if err != nil {
		return
	}
	res.NameLeft = fmt.Sprintf("%s [%s]", srcAddress, srcNetwork)
	res.NameRight = fmt.Sprintf("%s [%s]", destAddress, destNetwork)
	return
}
