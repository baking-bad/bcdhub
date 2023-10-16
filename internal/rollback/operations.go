package rollback

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func (rm Manager) rollbackOperations(ctx context.Context, level int64, rCtx *rollbackContext) error {
	log.Info().Msg("rollback operations...")

	ops, err := rm.rollback.GetOperations(ctx, level)
	if err != nil {
		return err
	}
	if len(ops) == 0 {
		return nil
	}

	for i := range ops {
		if ops[i].DestinationID > 0 {
			rCtx.applyOperationsCount(ops[i].DestinationID, 1)
			rCtx.applyTicketUpdates(ops[i].DestinationID, int64(ops[i].TicketUpdatesCount))
		}

		if ops[i].SourceID > 0 {
			rCtx.applyOperationsCount(ops[i].SourceID, 1)
		}

		switch ops[i].Kind {
		case types.OperationKindEvent:
			rCtx.generalStats.EventsCount -= 1
			rCtx.applyEvent(ops[i].SourceID)

		case types.OperationKindOrigination:
			rCtx.generalStats.OriginationsCount -= 1

		case types.OperationKindSrOrigination:
			rCtx.generalStats.SrOriginationsCount -= 1

		case types.OperationKindTransaction:
			rCtx.generalStats.TransactionsCount -= 1

		case types.OperationKindRegisterGlobalConstant:
			rCtx.generalStats.RegisterGlobalConstantCount -= 1

		case types.OperationKindSrExecuteOutboxMessage:
			rCtx.generalStats.SrExecutesCount -= 1

		case types.OperationKindTransferTicket:
			rCtx.generalStats.TransferTicketsCount -= 1
		}
	}

	count, err := rm.rollback.DeleteAll(ctx, (*operation.Operation)(nil), level)
	if err != nil {
		return errors.Wrap(err, "deleting operations")
	}
	rCtx.generalStats.OperationsCount -= count

	if err := rCtx.getLastActions(ctx, rm.rollback); err != nil {
		return errors.Wrap(err, "receiving last actions")
	}

	return nil
}
