package rollback

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/contract"
)

func (rm Manager) rollbackScripts(ctx context.Context, level int64) error {
	logger.Info().Msg("rollback scripts and global constants...")
	constants, err := rm.rollback.GlobalConstants(ctx, level)
	if err != nil {
		return err
	}
	scripts, err := rm.rollback.Scripts(ctx, level)
	if err != nil {
		return err
	}

	constantIds := make([]int64, len(constants))
	for i := range constants {
		constantIds[i] = constants[i].ID
	}
	scriptIds := make([]int64, len(scripts))
	for i := range scripts {
		scriptIds[i] = scripts[i].ID
	}

	if err := rm.rollback.DeleteScriptsConstants(ctx, scriptIds, constantIds); err != nil {
		return err
	}

	if len(scripts) > 0 {
		if _, err := rm.rollback.DeleteAll(ctx, (*contract.Script)(nil), level); err != nil {
			return err
		}
	}
	if len(constants) > 0 {
		if _, err := rm.rollback.DeleteAll(ctx, (*contract.GlobalConstant)(nil), level); err != nil {
			return err
		}
	}

	return nil
}
