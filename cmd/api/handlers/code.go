package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/formatter"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/macros"
	"github.com/gin-gonic/gin"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/tidwall/gjson"
)

// GetContractCode -
func (ctx *Context) GetContractCode(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	code, err := getContractCode(ctx.Dir, req.Network, req.Address)
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

	srcCode, err := getContractCode(ctx.Dir, req.SourceNetwork, req.SourceAddress)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	destCode, err := getContractCode(ctx.Dir, req.DestinationNetwork, req.DestinationAddress)
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

func getContractCode(dir, network, address string) (string, error) {
	filePath := fmt.Sprintf("%s/contracts/%s/%s.json", dir, network, address)
	if _, err := os.Stat(filePath); err != nil {
		return "", err
	}

	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	f.Close()

	contractJSON := gjson.ParseBytes(data).Get("script")
	collapsed, err := macros.FindMacros(contractJSON)
	if err != nil {
		return "", err
	}

	code := collapsed.Get("code")
	return formatter.MichelineToMichelson(code, false)
}
