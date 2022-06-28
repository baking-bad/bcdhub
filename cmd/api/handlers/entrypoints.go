package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/config"
	modelTypes "github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/gin-gonic/gin"
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
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/entrypoints [get]
func GetEntrypoints() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		script, err := getScript(ctx, req.Address, getSymLink(req.NetworkID()))
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		parameter, err := script.ParameterType()
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		entrypoints, err := parameter.GetEntrypointsDocs()
		if handleError(c, ctx.Storage, err, 0) {
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
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			resp[i].Schema = ast.WrapEntrypointJSONSchema(resp[i].Schema)
		}

		c.SecureJSON(http.StatusOK, resp)
	}
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
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/entrypoints/data [post]
func GetEntrypointData() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}
		var reqData getEntrypointDataRequest
		if err := c.BindJSON(&reqData); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		result, err := buildParametersForExecution(ctx, req.Address, getSymLink(req.NetworkID()), reqData.Name, reqData.Data)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		if reqData.Format == "michelson" {
			michelson, err := formatter.MichelineStringToMichelson(string(result.Value), false, formatter.DefLineSize)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			c.Data(http.StatusOK, gin.MIMEPlain, []byte(michelson))
			return
		}

		c.Data(http.StatusOK, gin.MIMEJSON, result.Value)
	}
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
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/entrypoints/schema [get]
func GetEntrypointSchema() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		var esReq entrypointSchemaRequest
		if err := c.BindQuery(&esReq); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		script, err := getScript(ctx, req.Address, getSymLink(req.NetworkID()))
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		parameter, err := script.ParameterType()
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		entrypoints, err := parameter.GetEntrypointsDocs()
		if handleError(c, ctx.Storage, err, 0) {
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
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			if esReq.FillType != "latest" {
				break
			}

			account, err := ctx.Accounts.Get(req.Address)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}

			op, err := ctx.Operations.Last(
				map[string]interface{}{
					"destination_id": account.ID,
					"kind":           modelTypes.OperationKindTransaction,
					"entrypoint":     esReq.EntrypointName,
					"status":         modelTypes.OperationStatusApplied,
				}, 0)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}

			if op.Parameters != nil {
				parameters := types.NewParameters(op.Parameters)
				subTree, err := parameter.FromParameters(parameters)
				if handleError(c, ctx.Storage, err, 0) {
					return
				}

				node, _ := subTree.UnwrapAndGetEntrypointName()
				schema.DefaultModel = make(ast.JSONModel)
				node.GetJSONModel(schema.DefaultModel)
			}
		}

		c.SecureJSON(http.StatusOK, schema)
	}
}

func buildParametersForExecution(ctx *config.Context, address, symLink, entrypoint string, data map[string]interface{}) (*types.Parameters, error) {
	parameterType, err := getParameterType(ctx, address, symLink)
	if err != nil {
		return nil, err
	}
	return parameterType.ParametersForExecution(entrypoint, data)
}
