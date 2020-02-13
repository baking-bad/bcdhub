package handlers

import (
	"errors"
	"net/http"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/formatter"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/macros"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
	"github.com/gin-gonic/gin"
	"github.com/pmezard/go-difflib/difflib"
)

// GetContractCode -
func (ctx *Context) GetContractCode(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	code, err := getContractCode(req.Network, req.Address, ctx.RPCs)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, code)
}

type getDiffRequest struct {
	SourceAddress      string `form:"sa"`
	SourceNetwork      string `form:"sn"`
	DestinationAddress string `form:"da"`
	DestinationNetwork string `form:"dn"`
}

// GetDiff -
func (ctx *Context) GetDiff(c *gin.Context) {
	var req getDiffRequest
	if err := c.BindQuery(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	srcCode, err := getContractCode(req.SourceNetwork, req.SourceAddress, ctx.RPCs)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	destCode, err := getContractCode(req.DestinationNetwork, req.DestinationAddress, ctx.RPCs)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(srcCode),
		B:        difflib.SplitLines(destCode),
		FromFile: req.SourceAddress,
		ToFile:   req.DestinationAddress,
		Context:  10,
	}
	text, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, text)
}

func getContractCode(network, address string, rpcs map[string]*noderpc.NodeRPC) (string, error) {
	rpc, ok := rpcs[network]
	if !ok {
		return "", errors.New("Unknown network")
	}

	contractJSON, err := rpc.GetScriptJSON(address, 0)
	if err != nil {
		return "", err
	}

	collapsed, err := macros.FindMacros(contractJSON)
	if err != nil {
		return "", err
	}

	code := collapsed.Get("code")
	return formatter.MichelineToMichelson(code, false)
}
