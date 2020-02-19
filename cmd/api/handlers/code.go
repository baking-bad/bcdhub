package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/formatter"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/macros"
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

	code, err := ctx.getContractCode(req.Network, req.Address)
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

	text, err := ctx.getDiff(req.SourceAddress, req.SourceNetwork, req.DestinationAddress, req.DestinationNetwork)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, text)
}

func (ctx *Context) getContractCode(network, address string) (string, error) {
	rpc, ok := ctx.RPCs[network]
	if !ok {
		return "", fmt.Errorf("Unknown network %s", network)
	}
	contract, err := contractparser.GetContract(rpc, address, network, 0, ctx.Dir)
	if err != nil {
		return "", err
	}

	contractJSON := contract.Get("script")
	collapsed, err := macros.FindMacros(contractJSON)
	if err != nil {
		return "", err
	}

	code := collapsed.Get("code")
	return formatter.MichelineToMichelson(code, false)
}

func (ctx *Context) getDiff(srcAddress, srcNetwork, destAddress, destNetwork string) (CodeDiff, error) {
	srcCode, err := ctx.getContractCode(srcNetwork, srcAddress)
	if err != nil {
		return CodeDiff{}, err
	}

	destCode, err := ctx.getContractCode(destNetwork, destAddress)
	if err != nil {
		return CodeDiff{}, err
	}

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(srcCode),
		B:        difflib.SplitLines(destCode),
		FromFile: srcAddress,
		ToFile:   destAddress,
		Context:  10,
	}
	text, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		return CodeDiff{}, err
	}

	buf := text
	buf = strings.ReplaceAll(buf, "+++", "+")
	buf = strings.ReplaceAll(buf, "++", "+")
	buf = strings.ReplaceAll(buf, "---", "-")
	buf = strings.ReplaceAll(buf, "--", "-")

	added := int64(strings.Count(buf, "+"))
	removed := int64(strings.Count(buf, "-"))

	return CodeDiff{
		Full:    text,
		Added:   added,
		Removed: removed,
	}, nil
}
