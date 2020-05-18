package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/docstring"
	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/jsonschema"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetEntrypoints godoc
// @Summary Get contract entrypoints
// @Description Get contract entrypoints
// @Tags contract
// @ID get-contract-entrypoints
// @Param network path string true "Network"
// @Param address path string true "KT address"
// @Accept json
// @Produce json
// @Success 200 {array} EntrypointSchema
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /contract/{network}/{address}/entrypoints [get]
func (ctx *Context) GetEntrypoints(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	metadata, err := getParameterMetadata(ctx.ES, req.Address, req.Network)
	if handleError(c, err, 0) {
		return
	}

	entrypoints, err := docstring.GetEntrypoints(metadata)
	if handleError(c, err, 0) {
		return
	}

	resp := make([]EntrypointSchema, len(entrypoints))
	for i, entrypoint := range entrypoints {
		resp[i].EntrypointType = entrypoint
		resp[i].Schema, resp[i].DefaultModel, err = jsonschema.Create(entrypoint.BinPath, metadata)
		if handleError(c, err, 0) {
			return
		}
	}

	c.JSON(http.StatusOK, resp)
}

// GetEntrypointData godoc
// @Summary Get entrypoint data from schema object
// @Description Get entrypoint data from schema object
// @Tags contract
// @ID get-contract-entrypoints-data
// @Param network path string true "Network"
// @Param address path string true "KT address"
// @Param bin_path body string true "Entrypoint binary path"
// @Param data body object true "Entrypoint`s arguments data"
// @Param format body object false "Return format (michelson | empty if micheline)"
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /contract/{network}/{address}/entrypoints/data [post]
func (ctx *Context) GetEntrypointData(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}
	var reqData getEntrypointDataRequest
	if err := c.BindJSON(&reqData); handleError(c, err, http.StatusBadRequest) {
		return
	}

	result, err := ctx.buildEntrypointMicheline(req.Network, req.Address, reqData.BinPath, reqData.Data)
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

func (ctx *Context) buildEntrypointMicheline(network, address, binPath string, data map[string]interface{}) (gjson.Result, error) {
	metadata, err := getParameterMetadata(ctx.ES, address, network)
	if err != nil {
		return gjson.Result{}, err
	}

	return metadata.BuildEntrypointMicheline(binPath, data)
}

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
