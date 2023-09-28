package handlers

import (
	"context"
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

		updates, err := ctx.TicketUpdates.Updates(c.Request.Context(), req.Address, args.Size, args.Offset)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		response, err := prepareTicketUpdates(c.Request.Context(), ctx, updates, nil)
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
		operation, err := ctx.Operations.GetByID(c.Request.Context(), req.ID)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		updates, err := ctx.TicketUpdates.UpdatesForOperation(c.Request.Context(), req.ID)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		response, err := prepareTicketUpdates(c.Request.Context(), ctx, updates, operation.Hash)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		c.SecureJSON(http.StatusOK, response)
	}
}

// GetTicketBalancesForAccount -
// @Router /v1/account/{network}/{address}/ticket_balances [get]
func GetTicketBalancesForAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("context").(*config.Context)

		var req getAccountRequest
		if err := c.BindUri(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, Error{Message: err.Error()})
			return
		}

		acc, err := ctx.Accounts.Get(c.Request.Context(), req.Address)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		balances, err := ctx.TicketUpdates.BalancesForAccount(c.Request.Context(), acc.ID, 10, 0)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		response := make([]TicketBalance, len(balances))
		for i := range balances {
			response[i] = NewTicketBalance(balances[i])
		}
		c.SecureJSON(http.StatusOK, response)
	}
}

func prepareTicketUpdates(c context.Context, ctx *config.Context, updates []ticket.TicketUpdate, hash []byte) ([]TicketUpdate, error) {
	response := make([]TicketUpdate, 0, len(updates))
	for i := range updates {
		update := NewTicketUpdateFromModel(updates[i])

		content, err := ast.NewTypedAstFromBytes(updates[i].Ticket.ContentType)
		if err != nil {
			return nil, err
		}
		docs, err := content.Docs("")
		if err != nil {
			return nil, err
		}
		update.ContentType = docs

		if err := content.SettleFromBytes(updates[i].Ticket.Content); err != nil {
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
			operation, err := ctx.Operations.GetByID(c, updates[i].OperationId)
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
