package handlers

import (
	"fmt"
	"net/http"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/formatter"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/macros"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
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

	d, err := ctx.getDiff(req.SourceAddress, req.SourceNetwork, req.DestinationAddress, req.DestinationNetwork, 0, 0)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, d)
}

func (ctx *Context) getContractCode(network, address string, level int64) (string, error) {
	contract, err := ctx.getContractCodeJSON(network, address, level)
	if err != nil {
		return "", err
	}

	code := contract.Get("code")
	return formatter.MichelineToMichelson(code, false)
}

func (ctx *Context) getContractCodeJSON(network, address string, level int64) (res gjson.Result, err error) {
	rpc, ok := ctx.RPCs[network]
	if !ok {
		return res, fmt.Errorf("Unknown network %s", network)
	}
	contract, err := contractparser.GetContract(rpc, address, network, level, ctx.Dir)
	if err != nil {
		return
	}

	contractJSON := contract.Get("script")
	return macros.FindMacros(contractJSON)
}

func (ctx *Context) getDiff(srcAddress, srcNetwork, destAddress, destNetwork string, levelSrc, levelDest int64) (res formatter.DiffResult, err error) {
	srcCode, err := ctx.getContractCodeJSON(srcNetwork, srcAddress, levelSrc)
	if err != nil {
		return
	}

	destCode, err := ctx.getContractCodeJSON(destNetwork, destAddress, levelDest)
	if err != nil {
		return
	}

	a := srcCode.Get("code")
	b := destCode.Get("code")
	res, err = formatter.Diff(a, b)
	if err != nil {
		return
	}
	res.NameA = fmt.Sprintf("%s [%s]", srcAddress, srcNetwork)
	res.NameB = fmt.Sprintf("%s [%s]", destAddress, destNetwork)
	return
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
