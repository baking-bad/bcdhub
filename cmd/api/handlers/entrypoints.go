package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
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
// @Router /v1/contract/{network}/{address}/entrypoints [get]
func (ctx *Context) GetEntrypoints(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	script, err := ctx.getScript(req.Address, req.Network, "")
	if ctx.handleError(c, err, 0) {
		return
	}
	parameter, err := script.ParameterType()
	if ctx.handleError(c, err, 0) {
		return
	}

	entrypoints, err := parameter.GetEntrypointsDocs()
	if ctx.handleError(c, err, 0) {
		return
	}

	resp := make([]EntrypointSchema, len(entrypoints))
	for i, entrypoint := range entrypoints {
		resp[i].EntrypointType = entrypoint
		e := parameter.FindByName(entrypoint.Name, true)
		if e == nil {
			continue
		}
		resp[i].Schema, err = e.ToJSONSchema()
		if ctx.handleError(c, err, 0) {
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
// @Param body body getEntrypointDataRequest true "Request body"
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/entrypoints/data [post]
func (ctx *Context) GetEntrypointData(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}
	var reqData getEntrypointDataRequest
	if err := c.BindJSON(&reqData); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	result, err := ctx.buildParametersForExecution(req.Network, req.Address, "", reqData.Name, reqData.Data)
	if ctx.handleError(c, err, 0) {
		return
	}

	if reqData.Format == "michelson" {
		michelson, err := formatter.MichelineStringToMichelson(string(result), false, formatter.DefLineSize)
		if ctx.handleError(c, err, 0) {
			return
		}
		c.JSON(http.StatusOK, michelson)
		return
	}

	c.JSON(http.StatusOK, result)
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
// @Success 200 {object} EntrypointSchema
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/entrypoints/schema [get]
func (ctx *Context) GetEntrypointSchema(c *gin.Context) {
	var req getContractRequest
	if err := c.BindUri(&req); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	var esReq entrypointSchemaRequest
	if err := c.BindQuery(&esReq); ctx.handleError(c, err, http.StatusBadRequest) {
		return
	}

	script, err := ctx.getScript(req.Address, req.Network, "")
	if ctx.handleError(c, err, 0) {
		return
	}
	parameter, err := script.ParameterType()
	if ctx.handleError(c, err, 0) {
		return
	}

	entrypoints, err := parameter.GetEntrypointsDocs()
	if ctx.handleError(c, err, 0) {
		return
	}

	schema := new(EntrypointSchema)
	for _, entrypoint := range entrypoints {
		if entrypoint.Name != esReq.EntrypointName {
			continue
		}

		schema.EntrypointType = entrypoint
		e := parameter.FindByName(esReq.EntrypointName, true)
		if e == nil {
			continue
		}
		schema.Schema, err = e.ToJSONSchema()
		if ctx.handleError(c, err, 0) {
			return
		}
		if esReq.FillType != "latest" {
			continue
		}

		op, err := ctx.Operations.Get(
			map[string]interface{}{
				"network":     req.Network,
				"destination": req.Address,
				"kind":        consts.Transaction,
				"entrypoint":  esReq.EntrypointName,
			},
			1,
			true,
		)
		if ctx.handleError(c, err, 0) {
			return
		}
		if len(op) != 1 {
			break
		}

		parameters := types.NewParameters([]byte(op[0].Parameters))
		var data ast.UntypedAST
		if err := json.Unmarshal(parameters.Value, &data); ctx.handleError(c, err, 0) {
			return
		}
		if err := e.ParseValue(data[0]); ctx.handleError(c, err, 0) {
			return
		}

		schema.DefaultModel = make(ast.JSONModel)
		e.GetJSONModel(schema.DefaultModel)
	}

	c.JSON(http.StatusOK, schema)
}

func (ctx *Context) buildParametersForExecution(network, address, protocol, entrypoint string, data map[string]interface{}) ([]byte, error) {
	parameterType, err := ctx.getParameterType(address, network, protocol)
	if err != nil {
		return nil, err
	}
	e := parameterType.FindByName(entrypoint, true)
	if e == nil {
		return nil, errors.Errorf("Unknown entrypoint name %s", entrypoint)
	}

	if err := e.FromJSONSchema(data); err != nil {
		return nil, err
	}

	return e.ToParameters()
}
