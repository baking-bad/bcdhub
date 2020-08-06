package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetDAppList -
func (ctx *Context) GetDAppList(c *gin.Context) {
	dapps, err := ctx.DB.GetDApps()
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, dapps)
}

// GetDApp -
func (ctx *Context) GetDApp(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	alias, err := ctx.DB.GetAlias(req.Address, req.Network)
	if handleError(c, err, 0) {
		return
	}

	dapp, err := ctx.DB.GetDApp(alias.DAppID)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, dapp)
}
