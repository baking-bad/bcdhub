package operations

import (
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/ticket"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
	"github.com/shopspring/decimal"
)

// TicketUpdateParser -
type TicketUpdateParser struct {
}

func (p *TicketUpdateParser) Parse(data *noderpc.OperationResult, operation *operation.Operation, store parsers.Store) {
	if data == nil {
		return
	}
	operation.TicketUpdates = make([]*ticket.TicketUpdate, 0)
	for i := range data.TicketUpdates {
		tckt := p.toModel(data.TicketUpdates[i], operation, store)
		operation.TicketUpdates = append(operation.TicketUpdates, tckt...)
	}
	for i := range data.TicketReceipt {
		tckt := p.toModel(data.TicketReceipt[i], operation, store)
		operation.TicketUpdates = append(operation.TicketUpdates, tckt...)
	}

	operation.TicketUpdatesCount = len(operation.TicketUpdates)
}

func (p *TicketUpdateParser) toModel(data noderpc.TicketUpdate, operation *operation.Operation, store parsers.Store) []*ticket.TicketUpdate {
	tckt := ticket.Ticket{
		ContentType: data.TicketToken.ContentType,
		Content:     data.TicketToken.Content,
		Ticketer: account.Account{
			Address:            data.TicketToken.Ticketer,
			Type:               types.NewAccountType(data.TicketToken.Ticketer),
			Level:              operation.Level,
			LastAction:         operation.Timestamp,
			TicketUpdatesCount: 1,
		},
		UpdatesCount: 1,
		Level:        operation.Level,
	}
	store.AddTickets(tckt)

	updates := make([]*ticket.TicketUpdate, 0)
	for i := range data.Updates {
		update := ticket.TicketUpdate{
			Level:     operation.Level,
			Timestamp: operation.Timestamp,
			Ticket:    tckt,
			Account: account.Account{
				Address:    data.Updates[i].Account,
				Type:       types.NewAccountType(data.Updates[i].Account),
				LastAction: operation.Timestamp,
				Level:      operation.Level,
			},
			Amount: decimal.RequireFromString(data.Updates[i].Amount),
		}
		updates = append(updates, &update)
		store.AddAccounts(update.Account, tckt.Ticketer)
		store.AddTicketBalances(ticket.Balance{
			Account: update.Account,
			Ticket:  tckt,
			Amount:  update.Amount.Copy(),
		})
	}
	return updates
}
