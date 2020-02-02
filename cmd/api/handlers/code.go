package handlers

import (
	"errors"
	"net/http"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/formatter"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/macros"
	"github.com/gin-gonic/gin"
)

// GetContractCode -
func (ctx *Context) GetContractCode(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	rpc, ok := ctx.RPCs[req.Network]
	if !ok {
		_ = c.AbortWithError(http.StatusBadRequest, errors.New("Unknown network"))
		return
	}

	contractJSON, err := rpc.GetScriptJSON(req.Address, 0)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	collapsed, err := macros.FindMacros(contractJSON)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	code := collapsed.Get("code")
	res := formatter.MichelineToMichelson(code, false)
	c.JSON(http.StatusOK, res)
}
