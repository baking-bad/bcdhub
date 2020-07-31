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
// @Param address path string true "KT address" minlength(36) maxlength(36)
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
// @Param address path string true "KT address" minlength(36) maxlength(36)
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

	result, err := ctx.buildEntrypointMicheline(req.Network, req.Address, reqData.BinPath, reqData.Data, false)
	if handleError(c, err, http.StatusBadRequest) {
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

// GetEntrypointSchema godoc
// @Summary Get contract`s entrypoint schema
// @Description Get contract`s entrypoint schema
// @Tags contract
// @ID get-contract-entrypoints-schema
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param entrypoint query string true "Entrypoint name"
// @Param fill_type query string false "Fill storage type" Enums(empty, latest)
// @Accept json
// @Produce json
// @Success 200 {array} EntrypointSchema
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /contract/{network}/{address}/entrypoints/schema [get]
func (ctx *Context) GetEntrypointSchema(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); handleError(c, err, http.StatusBadRequest) {
		return
	}

	var esReq entrypointSchemaRequest
	if err := c.BindQuery(&esReq); handleError(c, err, http.StatusBadRequest) {
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

	schema := new(EntrypointSchema)
	for _, entrypoint := range entrypoints {
		if entrypoint.Name != esReq.EntrypointName {
			continue
		}

		schema.EntrypointType = entrypoint
		schema.Schema, schema.DefaultModel, err = jsonschema.Create(entrypoint.BinPath, metadata)
		if handleError(c, err, 0) {
			return
		}
		if esReq.FillType != "latest" {
			continue
		}

		op, err := ctx.ES.GetOperations(
			map[string]interface{}{
				"network":     req.Network,
				"destination": req.Address,
				"kind":        consts.Transaction,
				"entrypoint":  esReq.EntrypointName,
			},
			true,
			true,
		)
		if handleError(c, err, 0) {
			return
		}
		if len(op) != 1 {
			break
		}
		parameters := gjson.Parse(op[0].Parameters)
		if parameters.Get("value").Exists() && parameters.Get("entrypoint").Exists() {
			parameters = parameters.Get("value")
		}
		if err := schema.DefaultModel.FillForEntrypoint(parameters, metadata, esReq.EntrypointName); handleError(c, err, 0) {
			return
		}
	}

	c.JSON(http.StatusOK, schema)
}

func (ctx *Context) buildEntrypointMicheline(network, address, binPath string, data map[string]interface{}, needValidate bool) (gjson.Result, error) {
	metadata, err := getParameterMetadata(ctx.ES, address, network)
	if err != nil {
		return gjson.Result{}, err
	}

	return metadata.BuildEntrypointMicheline(binPath, data, needValidate)
}

func getParameterMetadata(es elastic.IElastic, address, network string) (meta.Metadata, error) {
	state, err := es.GetLastBlock(network)
	if err != nil {
		return nil, err
	}

	metadata, err := meta.GetMetadata(es, address, consts.PARAMETER, state.Protocol)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func getStorageMetadata(es elastic.IElastic, address, network string) (meta.Metadata, error) {
	state, err := es.GetLastBlock(network)
	if err != nil {
		return nil, err
	}

	metadata, err := meta.GetMetadata(es, address, consts.STORAGE, state.Protocol)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}
