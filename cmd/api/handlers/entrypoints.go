package handlers

import (
	"context"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/encoding"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	modelTypes "github.com/baking-bad/bcdhub/internal/models/types"
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
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/entrypoints [get]
func GetEntrypoints() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.ShouldBindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		symLink, err := getCurrentSymLink(c.Request.Context(), ctx.Blocks)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		parameter, err := getParameterType(c.Request.Context(), ctx.Contracts, req.Address, symLink)
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
		if err := c.ShouldBindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}
		var reqData getEntrypointDataRequest
		if err := c.ShouldBindJSON(&reqData); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		symLink, err := getCurrentSymLink(c.Request.Context(), ctx.Blocks)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		result, err := buildParametersForExecution(c.Request.Context(), ctx, req.Address, symLink, reqData.Name, reqData.Data)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		if reqData.Format == "michelson" {
			michelson, err := formatter.MichelineStringToMichelson(string(result.Value), false, formatter.DefLineSize)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			c.SecureJSON(http.StatusOK, michelson)
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
		if err := c.ShouldBindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		var esReq entrypointSchemaRequest
		if err := c.ShouldBindQuery(&esReq); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		var (
			hash []byte
			err  error
		)

		if esReq.Hash != "" {
			hash, err = encoding.DecodeBase58(esReq.Hash)
			if handleError(c, ctx.Storage, err, http.StatusBadRequest) {
				return
			}
		}

		symLink, err := getCurrentSymLink(c.Request.Context(), ctx.Blocks)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		parameter, err := getParameterType(c.Request.Context(), ctx.Contracts, req.Address, symLink)
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

			var usingOperation operation.Operation
			switch esReq.FillType {
			case "latest":
				account, err := ctx.Accounts.Get(c.Request.Context(), req.Address)
				if handleError(c, ctx.Storage, err, 0) {
					return
				}

				op, err := ctx.Operations.Last(
					c.Request.Context(),
					map[string]interface{}{
						"destination_id": account.ID,
						"kind":           modelTypes.OperationKindTransaction,
						"entrypoint":     esReq.EntrypointName,
						"status":         modelTypes.OperationStatusApplied,
					}, 0)
				if handleError(c, ctx.Storage, err, 0) {
					return
				}
				usingOperation = op
			case "operation":
				if !bcd.IsOperationHash(esReq.Hash) || esReq.Counter == nil {
					handleError(c, ctx.Storage, errors.Errorf("invalid hash or counter for 'operation' type: hash=%s counter=%v", esReq.Hash, esReq.Counter), http.StatusBadRequest)
					return
				}

				opg, err := ctx.Operations.GetByHashAndCounter(c.Request.Context(), hash, int64(*esReq.Counter))
				if handleError(c, ctx.Storage, err, 0) {
					return
				}
				if len(opg) == 0 {
					break
				}

				usingOperation = opg[0]
			default:
				break
			}

			if usingOperation.Parameters != nil {
				parameters := types.NewParameters(usingOperation.Parameters)
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

func buildParametersForExecution(c context.Context, ctx *config.Context, address, symLink, entrypoint string, data map[string]interface{}) (*types.Parameters, error) {
	parameterType, err := getParameterType(c, ctx.Contracts, address, symLink)
	if err != nil {
		return nil, err
	}
	return parameterType.ParametersForExecution(entrypoint, data)
}
