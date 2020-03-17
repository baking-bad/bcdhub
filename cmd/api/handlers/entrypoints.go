package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/gin-gonic/gin"
)

// GetEntrypoints -
func (ctx *Context) GetEntrypoints(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	metadata, err := meta.GetMetadata(ctx.ES, req.Address, req.Network, "parameter", consts.HashBabylon)
	if handleError(c, err, 0) {
		return
	}

	entrypoints, err := metadata.GetEntrypoints()
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, entrypoints)
}
