package rollback

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/stats"
)

type rollbackContext struct {
	generalStats stats.Stats
	accountStats map[int64]*account.Account
}

func newRollbackContext(ctx context.Context, statsRepo stats.Repository) (rollbackContext, error) {
	stats, err := statsRepo.Get(ctx)
	if err != nil {
		return rollbackContext{}, err
	}
	return rollbackContext{
		generalStats: stats,
		accountStats: make(map[int64]*account.Account),
	}, nil
}

func (rCtx *rollbackContext) applyMigration(accountId int64) {
	if acc, ok := rCtx.accountStats[accountId]; ok {
		acc.MigrationsCount += 1
	} else {
		rCtx.accountStats[accountId] = &account.Account{
			ID:              accountId,
			MigrationsCount: 1,
		}
	}
}

func (rCtx *rollbackContext) applyEvent(accountId int64) {
	if acc, ok := rCtx.accountStats[accountId]; ok {
		acc.EventsCount += 1
	} else {
		rCtx.accountStats[accountId] = &account.Account{
			ID:          accountId,
			EventsCount: 1,
		}
	}
}

func (rCtx *rollbackContext) applyTicketUpdates(accountId int64, count int64) {
	if count < 1 {
		return
	}

	if acc, ok := rCtx.accountStats[accountId]; ok {
		acc.TicketUpdatesCount += count
	} else {
		rCtx.accountStats[accountId] = &account.Account{
			ID:                 accountId,
			TicketUpdatesCount: count,
		}
	}
}

func (rCtx *rollbackContext) applyOperationsCount(accountId int64, count int64) {
	if acc, ok := rCtx.accountStats[accountId]; ok {
		acc.OperationsCount += count
	} else {
		rCtx.accountStats[accountId] = &account.Account{
			ID:              accountId,
			OperationsCount: count,
		}
	}
}

func (rCtx *rollbackContext) getLastActions(ctx context.Context, rollback models.Rollback) error {
	addresses := make([]int64, 0, len(rCtx.accountStats))
	for _, acc := range rCtx.accountStats {
		addresses = append(addresses, acc.ID)
	}

	actions, err := rollback.GetLastAction(ctx, addresses...)
	if err != nil {
		return err
	}

	for i := range actions {
		if acc, ok := rCtx.accountStats[actions[i].AccountId]; ok {
			acc.LastAction = actions[i].Time
		} else {
			rCtx.accountStats[actions[i].AccountId] = &account.Account{
				ID:         actions[i].AccountId,
				LastAction: actions[i].Time,
			}
		}
	}
	return nil
}

func (rCtx *rollbackContext) update(ctx context.Context, rollback models.Rollback) error {
	if err := rollback.UpdateStats(ctx, rCtx.generalStats); err != nil {
		return err
	}

	for _, acc := range rCtx.accountStats {
		if err := rollback.UpdateAccountStats(ctx, *acc); err != nil {
			return err
		}
	}
	return nil
}
