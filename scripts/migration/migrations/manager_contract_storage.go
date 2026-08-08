package migrations

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// babylonActivationLevel is the level at which Babylon activated on mainnet,
// deprecating script-less "manager.tz" originations. Their storage never
// changes after origination, so reading it at any single later level works
// as a reference point for a one-off RPC read per contract.
const babylonActivationLevel = 655361

// babylonActivationTime is block 655361's timestamp. Every affected
// origination happened strictly before it, so it doubles as an upper bound
// for the query's time cursor, letting the planner prune hypertable chunks
// past this point instead of scanning years of unrelated history.
var babylonActivationTime = time.Date(2019, 10, 18, 8, 18, 28, 0, time.UTC)

// FixManagerContractStorage backfills `deffated_storage` for pre-Babylon
// "manager.tz" originations (no explicit `script`, only `managerPubkey`)
// that were indexed before Alpha.ParseOrigination started handling that case
// (see internal/parsers/storage/alpha.go) and were left with NULL storage.
type FixManagerContractStorage struct {
	bulkCount int
}

// Key -
func (m *FixManagerContractStorage) Key() string {
	return "fix_manager_contract_storage"
}

// Description -
func (m *FixManagerContractStorage) Description() string {
	return "backfill deffated_storage for pre-Babylon manager.tz originations left NULL by the old alpha parser"
}

// Do - migrate function
func (m *FixManagerContractStorage) Do(ctx *config.Context) error {
	if ctx.Network != types.Mainnet {
		log.Info().Str("network", ctx.Network.String()).Msg("fix_manager_contract_storage: skipping, bug is mainnet-only")
		return nil
	}

	if m.bulkCount == 0 {
		m.bulkCount = 200
	}

	background := context.Background()

	var (
		lastID        int64
		lastTimestamp time.Time
	)
	for {
		var ops []operation.Operation
		query := ctx.StorageDB.DB.NewSelect().
			Model(&ops).
			Relation("Destination").
			Where("operation.kind = 2").   // types.OperationKindOrigination
			Where("operation.status = 1"). // types.OperationStatusApplied
			Where("operation.deffated_storage IS NULL").
			Where("operation.timestamp < ?", babylonActivationTime)

		if !lastTimestamp.IsZero() {
			// `operation` is a Timescale hypertable partitioned by
			// timestamp: the cursor must include it (not just id) so the
			// planner can prune chunks instead of scanning all of them.
			query = query.Where("(operation.timestamp, operation.id) > (?, ?)", lastTimestamp, lastID)
		}

		if err := query.
			OrderExpr("operation.timestamp asc, operation.id asc").
			Limit(m.bulkCount).
			Scan(background); err != nil {
			return errors.Wrap(err, "select candidate originations")
		}

		if len(ops) == 0 {
			return nil
		}

		for i := range ops {
			address := ops[i].Destination.Address

			raw, err := ctx.RPC.GetScriptStorageRaw(background, address, babylonActivationLevel)
			if err != nil {
				return errors.Wrapf(err, "GetScriptStorageRaw %s", address)
			}
			if len(raw) == 0 {
				log.Warn().Str("address", address).Int64("operation_id", ops[i].ID).
					Msg("fix_manager_contract_storage: empty storage from node, skipping")
				continue
			}
			ops[i].DeffatedStorage = raw

			if _, err := ctx.StorageDB.DB.NewUpdate().
				Model(&ops[i]).
				Column("deffated_storage").
				WherePK().
				Exec(background); err != nil {
				return errors.Wrapf(err, "update operation id=%d", ops[i].ID)
			}
		}

		lastID = ops[len(ops)-1].ID
		lastTimestamp = ops[len(ops)-1].Timestamp
		log.Info().Int64("last_id", lastID).Time("last_timestamp", lastTimestamp).Msg("fix_manager_contract_storage: progress")
	}
}
