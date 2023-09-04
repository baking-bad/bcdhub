package handlers

import (
	"net/http"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/encoding"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/ticket"
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
		response, err := prepareTicketUpdates(ctx, updates, nil)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		c.SecureJSON(http.StatusOK, response)
	}
}

// GetTicketUpdatesForOperation -
// @Router /v1/operation/{network}/{id}/ticket_updates [get]
func GetTicketUpdatesForOperation() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getOperationByIDRequest
		if err := c.BindUri(&req); handleError(c, ctx.Storage, err, http.StatusBadRequest) {
			return
		}
		operation, err := ctx.Operations.GetByID(req.ID)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		updates, err := ctx.TicketUpdates.ForOperation(req.ID)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		response, err := prepareTicketUpdates(ctx, updates, operation.Hash)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		c.SecureJSON(http.StatusOK, response)
	}
}

func prepareTicketUpdates(ctx *config.Context, updates []ticket.TicketUpdate, hash []byte) ([]TicketUpdate, error) {
	response := make([]TicketUpdate, 0, len(updates))
	for i := range updates {
		update := NewTicketUpdateFromModel(updates[i])

		content, err := ast.NewTypedAstFromBytes(updates[i].ContentType)
		if err != nil {
			return nil, err
		}
		docs, err := content.Docs("")
		if err != nil {
			return nil, err
		}
		update.ContentType = docs

		if err := content.SettleFromBytes(updates[i].Content); err != nil {
			return nil, err
		}
		contentMiguel, err := content.ToMiguel()
		if err != nil {
			return nil, err
		}
		if len(contentMiguel) > 0 {
			update.Content = contentMiguel[0]
		}

		if len(hash) == 0 {
			operation, err := ctx.Operations.GetByID(updates[i].OperationID)
			if err != nil {
				return nil, err
			}
			update.OperationHash = encoding.MustEncodeOperationHash(operation.Hash)
		}
		if len(hash) > 0 {
			update.OperationHash = encoding.MustEncodeOperationHash(hash)
		}

		response = append(response, update)
	}

	return response, nil
}
