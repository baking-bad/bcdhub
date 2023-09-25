package rollback

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/logger"
)

func (rm Manager) rollbackBigMapState(ctx context.Context, level int64) error {
	logger.Info().Msg("rollback big map states...")
	states, err := rm.rollback.StatesChangedAtLevel(ctx, level)
	if err != nil {
		return err
	}

	for i, state := range states {
		diff, err := rm.rollback.LastDiff(ctx, state.Ptr, state.KeyHash, false)
		if err != nil {
			if rm.storage.IsRecordNotFound(err) {
				if err := rm.rollback.DeleteBigMapState(ctx, states[i]); err != nil {
					return err
				}
				continue
			}
			return err
		}
		states[i].LastUpdateLevel = diff.Level
		states[i].LastUpdateTime = diff.Timestamp
		states[i].IsRollback = true

		if len(diff.Value) > 0 {
			states[i].Value = diff.ValueBytes()
			states[i].Removed = false
		} else {
			states[i].Removed = true
			valuedDiff, err := rm.rollback.LastDiff(ctx, state.Ptr, state.KeyHash, true)
			if err != nil {
				if !rm.storage.IsRecordNotFound(err) {
					return err
				}
			} else {
				states[i].Value = valuedDiff.ValueBytes()
			}
		}

		if err := rm.rollback.SaveBigMapState(ctx, states[i]); err != nil {
			return err
		}
	}

	return nil
}
