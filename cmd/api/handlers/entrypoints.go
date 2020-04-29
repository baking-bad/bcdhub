package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/cmd/api/jsonschema"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/gin-gonic/gin"
)

func getParameterMetadata(es *elastic.Elastic, address, network string) (meta.Metadata, error) {
	state, err := es.CurrentState(network)
	if err != nil {
		return nil, err
	}

	metadata, err := meta.GetMetadata(es, address, consts.PARAMETER, state.Protocol)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

// GetEntrypoints -
func (ctx *Context) GetEntrypoints(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	metadata, err := getParameterMetadata(ctx.ES, req.Address, req.Network)
	if handleError(c, err, 0) {
		return
	}

	entrypoints, err := metadata.GetDocEntrypoints()
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

	metadata, err := getParameterMetadata(ctx.ES, req.Address, req.Network)
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

	metadata, err := getParameterMetadata(ctx.ES, req.Address, req.Network)
	if handleError(c, err, 0) {
		return
	}

	result, err := metadata.BuildEntrypointMicheline(reqData.BinPath, reqData.Data)
	if handleError(c, err, 0) {
		return
	}
	if reqData.Format == "michelson" {
		value := result.Get("value")
		michelson, err := formatter.MichelineToMichelson(value, false, formatter.DefLineSize)
		if handleError(c, err, 0) {
			return
		}
		c.JSON(http.StatusOK, michelson)
		return
	}
	c.JSON(http.StatusOK, result.Value())
}
