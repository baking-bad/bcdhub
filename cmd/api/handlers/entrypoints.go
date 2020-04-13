package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/cmd/api/jsonschema"
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

	metadata, err := meta.GetMetadata(ctx.ES, req.Address, consts.PARAMETER, consts.HashCarthage)
	if handleError(c, err, 0) {
		return
	}

	entrypoints, err := metadata.GetEntrypoints()
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, entrypoints)
}

// GetEntrypointSchema - returns entrypoint schema
func (ctx *Context) GetEntrypointSchema(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}
	var reqSchema getEntrypointSchemaRequest
	if err := c.BindQuery(&reqSchema); handleError(c, err, http.StatusBadRequest) {
		return
	}

	metadata, err := meta.GetMetadata(ctx.ES, req.Address, consts.PARAMETER, consts.HashCarthage)
	if handleError(c, err, 0) {
		return
	}

	schema, err := jsonschema.Create(reqSchema.BinPath, metadata)
	if handleError(c, err, 0) {
		return
	}

	c.JSON(http.StatusOK, schema)
}
