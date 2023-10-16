package handlers

import (
	"context"
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/encoding"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// GetOperation godoc
// @Summary Get operation group
// @Description Get operation group by hash
// @Tags operations
// @ID get-opg
// @Param network path string true "Network"
// @Param hash path string true "Operation group hash"  minlength(51) maxlength(51)
// @Param with_mempool query bool false "Search operation in mempool or not"
// @Param with_storage_diff query bool false "Include storage diff to operations or not"
// @Accept  json
// @Produce  json
// @Success 200 {array} Operation
// @Success 204 {object} gin.H
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/opg/{network}/{hash} [get]
func GetOperation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctxs := c.MustGet("contexts").(config.Contexts)
		any := ctxs.Any()

		var req OPGRequest
		if err := c.BindUri(&req); handleError(c, any.Storage, err, http.StatusBadRequest) {
			return
		}

		var queryReq opgRequest
		if err := c.BindQuery(&queryReq); handleError(c, any.Storage, err, http.StatusBadRequest) {
			return
		}

		hash, err := encoding.DecodeBase58(req.Hash)
		if handleError(c, any.Storage, err, http.StatusBadRequest) {
			return
		}

		network := req.NetworkID()
		if ctx, ok := ctxs[network]; ok {
			operations, err := ctx.Operations.GetByHash(c.Request.Context(), hash)
			if err != nil {
				handleError(c, ctx.Storage, err, 0)
				return
			}

			resp, err := PrepareOperations(c.Request.Context(), ctx, operations, queryReq.WithStorageDiff)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}

			c.SecureJSON(http.StatusOK, resp)
			return
		}

		for _, ctx := range ctxs {
			operations, err := ctx.Operations.GetByHash(c.Request.Context(), hash)
			if err != nil {
				handleError(c, ctx.Storage, err, 0)
				return
			}

			if len(operations) > 0 {
				resp, err := PrepareOperations(c.Request.Context(), ctx, operations, queryReq.WithStorageDiff)
				if handleError(c, ctx.Storage, err, 0) {
					return
				}

				c.SecureJSON(http.StatusOK, resp)
				return
			}
		}

		if !queryReq.WithMempool {
			c.SecureJSON(http.StatusNoContent, []gin.H{})
			return
		}

		operation, err := getOperationFromMempool(c, ctxs.Any(), req.Hash)
		if handleError(c, ctxs.Any().Storage, err, 0) {
			return
		}
		if operation != nil {
			c.SecureJSON(http.StatusOK, []Operation{*operation})
		} else {
			c.SecureJSON(http.StatusNoContent, []gin.H{})
		}
	}
}

// GetImplicitOperation godoc
// @Summary Get implicit operation
// @DescriptionGet implicit operation
// @Tags operations
// @ID get-implicit-operation
// @Param network path string true "Network"
// @Param counter path integer true "Counter"
// @Accept  json
// @Produce  json
// @Success 200 {array} Operation
// @Success 204 {object} gin.H
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/implicit/{network}/{counter} [get]
func GetImplicitOperation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req ImplicitOperationRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		operations, err := ctx.Operations.GetByHashAndCounter(c.Request.Context(), nil, req.Counter)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		resp, err := PrepareOperations(c.Request.Context(), ctx, operations, false)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, resp)
	}
}

// GetOperationErrorLocation godoc
// @Summary Get code line where operation failed
// @Description Get code line where operation failed
// @Tags operations
// @ID get-operation-error-location
// @Param network path string true "Network"
// @Param id path integer true "Internal BCD operation ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} GetErrorLocationResponse
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/operation/{network}/{id}/error_location [get]
func GetOperationErrorLocation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getOperationByIDRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		operation, err := ctx.Operations.GetByID(c.Request.Context(), req.ID)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		if !tezerrors.HasScriptRejectedError(operation.Errors) {
			handleError(c, ctx.Storage, errors.Errorf("No reject script error in operation"), http.StatusBadRequest)
			return
		}

		response, err := getErrorLocation(c.Request.Context(), ctx, operation, 2)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		c.SecureJSON(http.StatusOK, response)
	}
}

// GetOperationDiff godoc
// @Summary Get operation storage diff
// @DescriptionGet Get operation storage diff
// @Tags operations
// @ID get-operation-diff
// @Param network path string true "Network"
// @Param id path integer true "Internal BCD operation ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} ast.MiguelNode
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /v1/operation/{network}/{id}/diff [get]
func GetOperationDiff() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getOperationByIDRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}
		operation, err := ctx.Operations.GetByID(c.Request.Context(), req.ID)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		var result Operation
		result.FromModel(operation)

		if operation.CanHasStorageDiff() {
			proto, err := ctx.Cache.ProtocolByID(c.Request.Context(), operation.ProtocolID)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}

			if err := setFullStorage(c.Request.Context(), ctx, proto.SymLink, operation, &result); handleError(c, ctx.Storage, err, 0) {
				return
			}
		}
		c.SecureJSON(http.StatusOK, result.StorageDiff)
	}
}

// GetOperationGroups -
// @Summary Get operation groups by account
// @Description Get operation groups by account
// @Tags contract
// @ID get-operation-groups-by-account
// @Param network path string true "Network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param last_id query string false "Last operation ID"
// @Param size query integer false "Expected OPG count" mininum(1)
// @Accept  json
// @Produce  json
// @Success 200 {array} OPGResponse
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/opg [get]
func GetOperationGroups() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getAccountRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusNotFound) {
			return
		}

		var args opgForAddressRequest
		if err := c.BindQuery(&args); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		opg, err := ctx.Operations.OPG(c.Request.Context(), req.Address, int64(args.Size), args.LastID)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		response := make([]OPGResponse, len(opg))
		for i := range opg {
			response[i] = NewOPGResponse(opg[i])
		}
		c.SecureJSON(http.StatusOK, response)
	}
}

// GetByHashAndCounter -
// @Summary Get operations by hash and counter
// @Description Get operations by hash and counter
// @Tags operations
// @ID get-operations-by-hash-and-counter
// @Param network path string true "You can set network field for better performance"
// @Param hash path string true "Operation group hash"  minlength(51) maxlength(51)
// @Param counter path integer true "Counter of main operation"
// @Accept  json
// @Produce  json
// @Success 200 {array} Operation
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/opg/{network}/{hash}/{counter} [get]
func GetByHashAndCounter() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctxs := c.MustGet("contexts").(config.Contexts)

		var req OperationGroupContentRequest
		if err := c.BindUri(&req); handleError(c, ctxs.Any().Storage, err, http.StatusNotFound) {
			return
		}

		hash, err := encoding.DecodeBase58(req.Hash)
		if handleError(c, ctxs.Any().Storage, err, http.StatusBadRequest) {
			return
		}

		var opg []operation.Operation
		var foundContext *config.Context

		network := req.NetworkID()
		if ctx, ok := ctxs[network]; ok {
			opg, err = ctx.Operations.GetByHashAndCounter(c.Request.Context(), hash, req.Counter)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			foundContext = ctx
		} else {
			for _, ctx := range ctxs {
				opg, err = ctx.Operations.GetByHashAndCounter(c.Request.Context(), hash, req.Counter)
				if handleError(c, ctx.Storage, err, 0) {
					return
				}
				if len(opg) > 0 {
					foundContext = ctx
					break
				}
			}
		}

		resp, err := PrepareOperations(c.Request.Context(), foundContext, opg, false)
		if handleError(c, foundContext.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, resp)
	}
}

func getOperationFromMempool(c context.Context, ctx *config.Context, hash string) (*Operation, error) {
	res, err := ctx.Mempool.GetByHash(c, hash)
	if err != nil {
		return nil, err
	}

	switch {
	case len(res.Originations) > 0:
		return prepareMempoolOrigination(ctx, res.Originations[0]), nil
	case len(res.Transactions) > 0:
		return prepareMempoolTransaction(c, ctx, res.Transactions[0]), nil
	default:
		return nil, nil
	}
}
