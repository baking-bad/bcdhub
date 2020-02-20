package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/formatter"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/macros"
	"github.com/gin-gonic/gin"
	"github.com/pmezard/go-difflib/difflib"
)

type getContractCodeRequest struct {
	Address string `uri:"address"`
	Network string `uri:"network"`

	Level int64 `form:"level,omitempty"`
}

// GetContractCode -
func (ctx *Context) GetContractCode(c *gin.Context) {
	var req getContractCodeRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if err := c.BindQuery(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	code, err := ctx.getContractCode(req.Network, req.Address, req.Level)
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

	text, err := ctx.getDiff(req.SourceAddress, req.SourceNetwork, req.DestinationAddress, req.DestinationNetwork, 0, 0)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, text)
}

func (ctx *Context) getContractCode(network, address string, level int64) (string, error) {
	rpc, ok := ctx.RPCs[network]
	if !ok {
		return "", fmt.Errorf("Unknown network %s", network)
	}
	contract, err := contractparser.GetContract(rpc, address, network, level, ctx.Dir)
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

func (ctx *Context) getDiff(srcAddress, srcNetwork, destAddress, destNetwork string, levelSrc, levelDest int64) (CodeDiff, error) {
	srcCode, err := ctx.getContractCode(srcNetwork, srcAddress, levelSrc)
	if err != nil {
		return CodeDiff{}, err
	}

	destCode, err := ctx.getContractCode(destNetwork, destAddress, levelDest)
	if err != nil {
		return CodeDiff{}, err
	}

	nameSrc := srcAddress
	nameDest := destAddress
	if nameSrc == nameDest {
		nameSrc = fmt.Sprintf("%s before babylon", nameSrc)
		nameDest = fmt.Sprintf("%s after babylon", nameDest)
	}

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(srcCode),
		B:        difflib.SplitLines(destCode),
		FromFile: nameSrc,
		ToFile:   nameDest,
		Context:  5,
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

// GetMigrationDiff -
func (ctx *Context) GetMigrationDiff(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	contract, err := ctx.ES.GetContract(map[string]interface{}{
		"address": req.Address,
		"network": consts.Mainnet,
	})
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if contract.Level >= consts.LevelBabylon {
		_ = c.AbortWithError(http.StatusBadRequest, fmt.Errorf("No migrations for contract %s", req.Address))
		return
	}

	codeDiff, err := ctx.getDiff(contract.Address, contract.Network, contract.Address, contract.Network, contract.Level, 0)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, codeDiff)
}
