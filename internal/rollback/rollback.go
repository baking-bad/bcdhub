package rollback

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	smartrollup "github.com/baking-bad/bcdhub/internal/models/smart_rollup"
	"github.com/baking-bad/bcdhub/internal/models/ticket"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/pkg/errors"
)

// Manager -
type Manager struct {
	storage   models.GeneralRepository
	blockRepo block.Repository
	rollback  models.Rollback
}

// NewManager -
func NewManager(storage models.GeneralRepository, blockRepo block.Repository, saver models.Rollback) Manager {
	return Manager{
		storage, blockRepo, saver,
	}
}

// Rollback - rollback indexer state to level
func (rm Manager) Rollback(ctx context.Context, network types.Network, fromState block.Block, toLevel int64) error {
	if toLevel >= fromState.Level {
		return errors.Errorf("To level must be less than from level: %d >= %d", toLevel, fromState.Level)
	}

	for level := fromState.Level; level > toLevel; level-- {
		logger.Info().Str("network", network.String()).Msgf("Rollback to %d block", level)

		if _, err := rm.blockRepo.Get(ctx, level); err != nil {
			if rm.storage.IsRecordNotFound(err) {
				continue
			}
			return err
		}

		if err := rm.rollbackBlock(ctx, level); err != nil {
			logger.Error().Err(err).Str("network", network.String()).Msg("rollback error")
			return rm.rollback.Rollback()
		}
	}

	return rm.rollback.Commit()
}

func (rm Manager) rollbackBlock(ctx context.Context, level int64) error {
	if err := rm.rollbackOperations(ctx, level); err != nil {
		return err
	}
	if err := rm.rollbackBigMapState(ctx, level); err != nil {
		return err
	}
	if err := rm.rollbackScripts(ctx, level); err != nil {
		return err
	}
	if err := rm.rollbackAll(ctx, level); err != nil {
		return err
	}
	if err := rm.rollback.Protocols(ctx, level); err != nil {
		return err
	}
	return nil
}

func (rm Manager) rollbackAll(ctx context.Context, level int64) error {
	for _, model := range []models.Model{
		(*block.Block)(nil),
		(*contract.Contract)(nil),
		(*bigmapdiff.BigMapDiff)(nil),
		(*bigmapaction.BigMapAction)(nil),
		(*smartrollup.SmartRollup)(nil),
		(*ticket.TicketUpdate)(nil),
		(*migration.Migration)(nil),
		(*account.Account)(nil),
	} {
		if err := rm.rollback.DeleteAll(ctx, model, level); err != nil {
			return err
		}
		logger.Info().
			Msgf("rollback: %T", model)
	}
	return nil
}
