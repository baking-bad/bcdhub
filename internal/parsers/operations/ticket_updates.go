package operations

import (
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/ticket"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/shopspring/decimal"
)

// TicketUpdateParser -
type TicketUpdateParser struct {
}

func (p *TicketUpdateParser) Parse(data *noderpc.OperationResult, operation *operation.Operation) {
	if data == nil {
		return
	}
	operation.TickerUpdates = make([]*ticket.TicketUpdate, 0)
	for i := range data.TicketUpdates {
		tckt := p.toModel(data.TicketUpdates[i], operation)
		operation.TickerUpdates = append(operation.TickerUpdates, tckt...)
	}
	for i := range data.TicketReceipt {
		tckt := p.toModel(data.TicketReceipt[i], operation)
		operation.TickerUpdates = append(operation.TickerUpdates, tckt...)
	}

	operation.TicketUpdatesCount = len(operation.TickerUpdates)
}

func (p *TicketUpdateParser) toModel(data noderpc.TicketUpdate, operation *operation.Operation) []*ticket.TicketUpdate {
	updates := make([]*ticket.TicketUpdate, 0)
	for i := range data.Updates {
		updates = append(updates, &ticket.TicketUpdate{
			Level:     operation.Level,
			Timestamp: operation.Timestamp,
			Ticketer: account.Account{
				Address: data.TicketToken.Ticketer,
				Type:    types.NewAccountType(data.TicketToken.Ticketer),
			},
			ContentType: data.TicketToken.ContentType,
			Content:     data.TicketToken.Content,
			Account: account.Account{
				Address: data.Updates[i].Account,
				Type:    types.NewAccountType(data.Updates[i].Account),
			},
			Amount: decimal.RequireFromString(data.Updates[i].Amount),
		})
	}
	return updates
}
