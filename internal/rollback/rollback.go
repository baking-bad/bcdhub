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
	"github.com/baking-bad/bcdhub/internal/models/stats"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/pkg/errors"
)

// Manager -
type Manager struct {
	storage   models.GeneralRepository
	blockRepo block.Repository
	rollback  models.Rollback
	statsRepo stats.Repository
}

// NewManager -
func NewManager(
	storage models.GeneralRepository,
	blockRepo block.Repository,
	rollback models.Rollback,
	statsRepo stats.Repository,
) Manager {
	return Manager{
		storage:   storage,
		blockRepo: blockRepo,
		rollback:  rollback,
		statsRepo: statsRepo,
	}
}

// Rollback - rollback indexer state to level
func (rm Manager) Rollback(ctx context.Context, network types.Network, fromState block.Block, toLevel int64) error {
	if toLevel >= fromState.Level {
		return errors.Errorf("To level must be less than from level: %d >= %d", toLevel, fromState.Level)
	}

	for level := fromState.Level; level > toLevel; level-- {
		logger.Info().Str("network", network.String()).Msgf("start rollback to %d", level)

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

		logger.Info().Str("network", network.String()).Msgf("rolled back to %d", level)
	}

	return rm.rollback.Commit()
}

func (rm Manager) rollbackBlock(ctx context.Context, level int64) error {
	rollbackCtx, err := newRollbackContext(ctx, rm.statsRepo)
	if err != nil {
		return err
	}

	if err := rm.rollbackOperations(ctx, level, &rollbackCtx); err != nil {
		return err
	}
	if err := rm.rollbackBigMapState(ctx, level); err != nil {
		return err
	}
	if err := rm.rollbackScripts(ctx, level); err != nil {
		return err
	}
	if err := rm.rollbackMigrations(ctx, level, &rollbackCtx); err != nil {
		return err
	}
	if err := rm.rollbackTickets(ctx, level); err != nil {
		return err
	}
	if err := rm.rollbackAll(ctx, level, &rollbackCtx); err != nil {
		return err
	}
	if err := rm.rollback.Protocols(ctx, level); err != nil {
		return err
	}
	if err := rollbackCtx.update(ctx, rm.rollback); err != nil {
		return err
	}

	return nil
}

func (rm Manager) rollbackMigrations(ctx context.Context, level int64, rCtx *rollbackContext) error {
	migrations, err := rm.rollback.GetMigrations(ctx, level)
	if err != nil {
		return nil
	}
	if len(migrations) == 0 {
		return nil
	}

	for i := range migrations {
		rCtx.applyMigration(migrations[i].Contract.AccountID)
	}

	if _, err := rm.rollback.DeleteAll(ctx, (*migration.Migration)(nil), level); err != nil {
		return err
	}
	logger.Info().Msg("rollback migrations")
	return nil
}

func (rm Manager) rollbackAll(ctx context.Context, level int64, rCtx *rollbackContext) error {
	for _, model := range []models.Model{
		(*block.Block)(nil),
		(*bigmapdiff.BigMapDiff)(nil),
		(*bigmapaction.BigMapAction)(nil),
		(*smartrollup.SmartRollup)(nil),
		(*account.Account)(nil),
	} {
		if _, err := rm.rollback.DeleteAll(ctx, model, level); err != nil {
			return err
		}
		logger.Info().Msgf("rollback: %T", model)
	}

	contractsCount, err := rm.rollback.DeleteAll(ctx, (*contract.Contract)(nil), level)
	if err != nil {
		return err
	}
	rCtx.generalStats.ContractsCount -= contractsCount
	logger.Info().Msgf("rollback contracts")

	return nil
}
