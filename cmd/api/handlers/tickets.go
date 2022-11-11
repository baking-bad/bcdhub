package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/gin-gonic/gin"
)

// GetContractTicketUpdates godoc
// @Summary Get ticket updates for contract
// @Description Get ticket updates for contract
// @Tags contract
// @ID get-contract-ticket-updates
// @Param network path string true "network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param size query integer false "Updates count" mininum(1) maximum(10)
// @Param offset query integer false "Offset" mininum(1)
// @Accept json
// @Produce json
// @Success 200 {array} GlobalConstant
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/ticket_updates [get]
func GetContractTicketUpdates() gin.HandlerFunc {
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

		updates, err := ctx.TicketUpdates.Get(req.Address, args.Size, args.Offset)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		response := make([]TicketUpdate, 0, len(updates))
		for i := range updates {
			update := NewTicketUpdateFromModel(updates[i])

			content, err := ast.NewTypedAstFromBytes(updates[i].ContentType)
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			docs, err := content.Docs("")
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			update.ContentType = docs

			if err := content.SettleFromBytes(updates[i].Content); handleError(c, ctx.Storage, err, 0) {
				return
			}
			contentMiguel, err := content.ToMiguel()
			if handleError(c, ctx.Storage, err, 0) {
				return
			}
			if len(contentMiguel) > 0 {
				update.Content = contentMiguel[0]
			}

			response = append(response, update)
		}

		c.SecureJSON(http.StatusOK, response)
	}
}
