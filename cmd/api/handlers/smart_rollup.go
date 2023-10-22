package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/gin-gonic/gin"
)

// GetSmartRollup godoc
// @Summary Get smart rollup
// @Description Get smart rollup
// @Tags smart-rollups
// @ID get-smart-rollups
// @Param network path string true "network"
// @Param address path string true "expr address of smart rollup" minlength(36) maxlength(36)
// @Accept json
// @Produce json
// @Success 200 {object} SmartRollup
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/smart_rollups/{network}/{address} [get]
func GetSmartRollup() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getSmartRollupRequest
		if err := c.ShouldBindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		rollup, err := ctx.SmartRollups.Get(c.Request.Context(), req.Address)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		response := NewSmartRollup(rollup)
		typ, err := ast.NewTypedAstFromBytes(rollup.Type)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		docs, err := typ.Docs("")
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		response.Type = docs
		c.SecureJSON(http.StatusOK, response)
	}
}

// ListSmartRollups godoc
// @Summary List smart rollups
// @Description List smart rollups
// @Tags smart-rollups
// @ID list-smart-rollups
// @Param network path string true "network"
// @Param size query integer false "Constants count" mininum(1) maximum(10)
// @Param offset query integer false "Offset" mininum(1)
// @Param sort query string false "Sort order" Enums(asc, desc)
// @Accept json
// @Produce json
// @Success 200 {array} SmartRollup
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/smart_rollups/{network} [get]
func ListSmartRollups() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var args smartRollupListRequest
		if err := c.ShouldBindQuery(&args); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}

		rollups, err := ctx.SmartRollups.List(c.Request.Context(), args.Size, args.Offset, args.Sort)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		response := make([]SmartRollup, len(rollups))
		for i := range rollups {
			response[i] = NewSmartRollup(rollups[i])

			typ, err := ast.NewTypedAstFromBytes(rollups[i].Type)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			docs, err := typ.Docs("")
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			response[i].Type = docs
		}
		c.SecureJSON(http.StatusOK, response)
	}
}
