package rollback

import (
	"context"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models/ticket"
)

func (rm Manager) rollbackTicketUpdates(ctx context.Context, level int64) error {
	updates, err := rm.rollback.GetTicketUpdates(ctx, level)
	if err != nil {
		return err
	}

	if len(updates) == 0 {
		return nil
	}

	balances := make(map[string]*ticket.Balance)
	for i := range updates {
		key := fmt.Sprintf("%d_%d", updates[i].AccountId, updates[i].TicketId)
		if b, ok := balances[key]; ok {
			b.Amount = b.Amount.Add(updates[i].Amount)
		} else {
			balances[key] = &ticket.Balance{
				AccountId: updates[i].AccountId,
				TicketId:  updates[i].TicketId,
				Amount:    updates[i].Amount.Copy(),
			}
		}
	}

	arr := make([]*ticket.Balance, 0, len(balances))
	for _, balance := range balances {
		arr = append(arr, balance)
	}

	return rm.rollback.TicketBalances(ctx, arr...)
}

func (rm Manager) rollbackTickets(ctx context.Context, level int64) error {
	if err := rm.rollbackTicketUpdates(ctx, level); err != nil {
		return err
	}

	if _, err := rm.rollback.DeleteAll(ctx, (*ticket.TicketUpdate)(nil), level); err != nil {
		return err
	}

	ticketsIds, err := rm.rollback.DeleteTickets(ctx, level)
	if err != nil {
		return err
	}
	if len(ticketsIds) == 0 {
		return nil
	}

	return rm.rollback.DeleteTicketBalances(ctx, ticketsIds)
}
