package rollback

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/operation"
)

func (rm Manager) rollbackOperations(ctx context.Context, level int64) error {
	logger.Info().Msg("rollback operations...")

	ops, err := rm.rollback.GetOperations(ctx, level)
	if err != nil {
		return err
	}
	if len(ops) == 0 {
		return nil
	}

	if err := rm.rollback.DeleteAll(ctx, (*operation.Operation)(nil), level); err != nil {
		return err
	}

	accounts := make(map[int64]int64)
	for i := range ops {
		if !ops[i].Destination.IsEmpty() {
			if _, ok := accounts[ops[i].DestinationID]; !ok {
				accounts[ops[i].DestinationID] = 1
			} else {
				accounts[ops[i].DestinationID] += 1
			}
		}

		if !ops[i].Source.IsEmpty() {
			if _, ok := accounts[ops[i].SourceID]; !ok {
				accounts[ops[i].SourceID] = 1
			} else {
				accounts[ops[i].SourceID] += 1
			}
		}
	}

	if len(accounts) == 0 {
		return nil
	}

	addresses := make([]int64, 0, len(accounts))
	for address := range accounts {
		addresses = append(addresses, address)
	}

	actions, err := rm.rollback.GetLastAction(ctx, addresses...)
	if err != nil {
		return err
	}

	for i := range actions {
		count, ok := accounts[actions[i].AccountId]
		if !ok {
			count = 1
		}

		if err := rm.rollback.UpdateAccountStats(ctx, actions[i].AccountId, actions[i].Time, count); err != nil {
			return err
		}
	}

	return nil
}
