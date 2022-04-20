package migrations

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/go-pg/pg/v10"
)

// BigMapStateCount -
type BigMapStateCount struct{}

// Key -
func (m *BigMapStateCount) Key() string {
	return "big_map_state_count"
}

// Description -
func (m *BigMapStateCount) Description() string {
	return "set big map state count"
}

// Do - migrate function
func (m *BigMapStateCount) Do(ctx *config.Context) error {
	var offset int
	var end bool
	for !end {
		if err := ctx.StorageDB.DB.RunInTransaction(context.Background(), func(tx *pg.Tx) error {
			var states []bigmapdiff.BigMapState
			if err := tx.Model(&bigmapdiff.BigMapState{}).Order("id asc").Limit(10000).Offset(offset).Select(&states); err != nil {
				return err
			}

			for _, state := range states {
				count, err := tx.Model(&bigmapdiff.BigMapDiff{}).
					Where("ptr = ?", state.Ptr).Where("key_hash = ?", state.KeyHash).Where("contract = ?", state.Contract).
					Count()
				if err != nil {
					return err
				}
				state.Count = int64(count)

				if _, err := tx.Model(&state).Set("count = ?count").Where("id = ?id").Update(); err != nil {
					return err
				}
			}

			offset += len(states)
			end = len(states)%10000 != 0

			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}
