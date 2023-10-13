package handlers

import (
	"context"
	"net/http"

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
// @Success 200 {array} TicketUpdate
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

		updates, err := ctx.Tickets.Updates(c.Request.Context(), req.Address, args.Size, args.Offset)
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

// GetContractTickets godoc
// @Summary Get tickets for contract
// @Description Get tickets for contract
// @Tags contract
// @ID get-contract-tickets
// @Param network path string true "network"
// @Param address path string true "KT address" minlength(36) maxlength(36)
// @Param size query integer false "Updates count" mininum(1) maximum(10)
// @Param offset query integer false "Offset" mininum(1)
// @Accept json
// @Produce json
// @Success 200 {array} Ticket
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /v1/contract/{network}/{address}/tickets [get]
func GetContractTickets() gin.HandlerFunc {
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

		tickets, err := ctx.Tickets.List(c.Request.Context(), req.Address, args.Size, args.Offset)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		response, err := prepareTickets(tickets)
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
		updates, err := ctx.Tickets.UpdatesForOperation(c.Request.Context(), req.ID)
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

		var args ticketBalancesRequest
		if err := c.BindQuery(&args); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, Error{Message: err.Error()})
			return
		}

		acc, err := ctx.Accounts.Get(c.Request.Context(), req.Address)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}

		balances, err := ctx.Tickets.BalancesForAccount(c.Request.Context(), acc.ID, ticket.BalanceRequest{
			Limit:               args.Size,
			Offset:              args.Offset,
			WithoutZeroBalances: args.WithoutZeroBalances,
		})
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		response, err := prepareTicketBalances(balances)
		if handleError(c, ctx.Storage, err, 0) {
			return
		}
		c.SecureJSON(http.StatusOK, response)
	}
}

func prepareTicketUpdates(c context.Context, ctx *config.Context, updates []ticket.TicketUpdate, hash []byte) ([]TicketUpdate, error) {
	response := make([]TicketUpdate, 0, len(updates))
	for i := range updates {
		update := NewTicketUpdateFromModel(updates[i])
		ticket, err := NewTicket(updates[i].Ticket)
		if err != nil {
			return nil, err
		}
		update.ContentType = ticket.ContentType
		update.Content = ticket.Content

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

func prepareTicketBalances(balances []ticket.Balance) ([]TicketBalance, error) {
	response := make([]TicketBalance, len(balances))
	for i := range balances {
		balance := NewTicketBalance(balances[i])
		ticket, err := NewTicket(balances[i].Ticket)
		if err != nil {
			return nil, err
		}
		balance.ContentType = ticket.ContentType
		balance.Content = ticket.Content
		response[i] = balance
	}
	return response, nil
}

func prepareTickets(tickets []ticket.Ticket) ([]Ticket, error) {
	response := make([]Ticket, len(tickets))
	for i := range tickets {
		ticket, err := NewTicket(tickets[i])
		if err != nil {
			return nil, err
		}
		response[i] = ticket
	}
	return response, nil
}
