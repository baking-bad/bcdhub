package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetGlobalConstant godoc
// @Summary Get global constant
// @Description Get global constant
// @Tags global-constants
// @ID get-global-constant
// @Param network path string true "network"
// @Param address path string true "expr address of constant" minlength(54) maxlength(54)
// @Accept json
// @Produce json
// @Success 200 {object} GlobalConstant
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/global_constants/{network}/{address} [get]
func GetGlobalConstant() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getGlobalConstantRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		constant, err := ctx.GlobalConstants.Get(c.Request.Context(), req.Address)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		michelson, err := formatter.MichelineToMichelson(gjson.ParseBytes(constant.Value), false, formatter.DefLineSize)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		globalConstant := NewGlobalConstantFromModel(constant)
		globalConstant.Code = michelson

		c.SecureJSON(http.StatusOK, globalConstant)
	}
}

// ListGlobalConstants godoc
// @Summary List global constants
// @Description List global constants
// @Tags global-constants
// @ID list-global-constants
// @Param network path string true "network"
// @Param size query integer false "Constants count" mininum(1) maximum(10)
// @Param offset query integer false "Offset" mininum(1)
// @Param order_by query string false "Order by" Enums(level, timestamp, links_count, address)
// @Param sort query string false "Sort order" Enums(asc, desc)
// @Accept json
// @Produce json
// @Success 200 {array} contract.ListGlobalConstantItem
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/global_constants/{network} [get]
func ListGlobalConstants() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var args globalConstantsListRequest
		if err := c.BindQuery(&args); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		constants, err := ctx.GlobalConstants.List(c.Request.Context(), args.Size, args.Offset, args.OrderBy, args.Sort)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		c.SecureJSON(http.StatusOK, constants)
	}
}

// GetContractGlobalConstants godoc
// @Summary Get global constants used by contract
// @Description Get global constants used by contract
// @Tags contract
// @ID get-contract-global-constants
// @Param network path string true "network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param size query integer false "Constants count" mininum(1) maximum(10)
// @Param offset query integer false "Offset" mininum(1)
// @Accept json
// @Produce json
// @Success 200 {array} GlobalConstant
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/global_constants [get]
func GetContractGlobalConstants() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getContractRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		var args pageableRequest
		if err := c.BindQuery(&args); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		constants, err := ctx.GlobalConstants.ForContract(c.Request.Context(), req.Address, args.Size, args.Offset)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		response := make([]GlobalConstant, 0, len(constants))
		for i := range constants {
			response = append(response, NewGlobalConstantFromModel(constants[i]))
		}

		c.SecureJSON(http.StatusOK, response)
	}
}

// GetGlobalConstantContracts godoc
// @Summary Get contracts that use the global constant
// @Description Get contracts that use the global constant
// @Tags global-constants
// @ID get-global-constant-contracts
// @Param network path string true "network"
// @Param address path string true "expr address of constant" minlength(54) maxlength(54)
// @Param size query integer false "Constants count" mininum(1) maximum(10)
// @Param offset query integer false "Offset" mininum(1)
// @Accept json
// @Produce json
// @Success 200 {array} Contract
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/global_constants/{network}/{address}/contracts [get]
func GetGlobalConstantContracts() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var args globalConstantsContractsRequest
		if err := c.BindUri(&args); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		if err := c.BindQuery(&args); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		contracts, err := ctx.GlobalConstants.ContractList(c.Request.Context(), args.Address, args.Size, args.Offset)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		response := make([]Contract, 0, len(contracts))
		for i := range contracts {
			item, err := contractPostprocessing(ctx, contracts[i])
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			response = append(response, item)
		}

		c.SecureJSON(http.StatusOK, response)
	}
}
