package rollback

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
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

	contracts := make(map[int64]int64)
	for i := range ops {
		if ops[i].IsOrigination() {
			continue
		}
		if ops[i].Destination.Type == types.AccountTypeContract {
			if _, ok := contracts[ops[i].DestinationID]; !ok {
				contracts[ops[i].DestinationID] = 1
			} else {
				contracts[ops[i].DestinationID] += 1
			}
		}
		if ops[i].Source.Type == types.AccountTypeContract {
			if _, ok := contracts[ops[i].SourceID]; !ok {
				contracts[ops[i].SourceID] = 1
			} else {
				contracts[ops[i].SourceID] += 1
			}
		}
	}

	if len(contracts) == 0 {
		return nil
	}

	addresses := make([]int64, 0, len(contracts))
	for address := range contracts {
		addresses = append(addresses, address)
	}

	actions, err := rm.rollback.GetContractsLastAction(ctx, addresses...)
	if err != nil {
		return err
	}

	for i := range actions {
		count, ok := contracts[actions[i].AccountId]
		if !ok {
			count = 1
		}

		if err := rm.rollback.UpdateContractStats(ctx, actions[i].AccountId, actions[i].Time, count); err != nil {
			return err
		}
	}

	return nil
}
