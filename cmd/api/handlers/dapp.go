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
	var req getDappRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	dapp, err := ctx.DB.GetDAppBySlug(req.Slug)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, dapp)
}
