package handlers

import (
	"net/http"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/miguel"
	"github.com/gin-gonic/gin"
)

// GetEntrypoints -
func (ctx *Context) GetEntrypoints(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	metadata, err := getMetadata(ctx.ES, req.Address, req.Network, "parameter", 0)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	entrypoints, err := miguel.GetEntrypoints(metadata)
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, entrypoints)
}
