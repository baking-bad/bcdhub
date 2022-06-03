package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/gin-gonic/gin"
)

// GetGlobalConstant godoc
// @Summary Get global constant
// @Description Get global constant
// @Tags contract
// @ID get-global-constant
// @Param network path string true "network"
// @Param address path string true "expr address of constant" minlength(54) maxlength(54)
// @Accept json
// @Produce json
// @Success 200 {object} GlobalConstant
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

		constant, err := ctx.GlobalConstants.Get(req.Address)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		c.SecureJSON(http.StatusOK, NewGlobalConstantFromModel(constant))
	}
}

// ListGlobalConstants godoc
// @Summary List global constants
// @Description List global constants
// @Tags contract
// @ID list-global-constants
// @Param network path string true "network"
// @Param size query integer false "Constants count" mininum(1) maximum(10)
// @Param offset query integer false "Offset" mininum(1)
// @Accept json
// @Produce json
// @Success 200 {array} GlobalConstant
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/global_constants/{network} [get]
func ListGlobalConstants() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var args pageableRequest
		if err := c.BindQuery(&args); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		constants, err := ctx.GlobalConstants.List(args.Size, args.Offset)
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
