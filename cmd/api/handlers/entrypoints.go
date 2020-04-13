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

// GetEntrypointData - returns entrypoint data from schema object
func (ctx *Context) GetEntrypointData(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}
	var reqData getEntrypointDataRequest
	if err := c.BindJSON(&reqData); handleError(c, err, http.StatusBadRequest) {
		return
	}

	metadata, err := meta.GetMetadata(ctx.ES, req.Address, consts.PARAMETER, consts.HashCarthage)
	if handleError(c, err, 0) {
		return
	}

	result, err := metadata.BuildEntrypointMicheline(reqData.BinPath, reqData.Data)
	if handleError(c, err, 0) {
		return
	}
	c.JSON(http.StatusOK, result)
}
