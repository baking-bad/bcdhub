package migrations

import (
	"context"
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	bcdTypes "github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

// oldUnnamedEntrypointPattern matches names stored by the old `unwrap`: `@entrypoint_K`,
// where K is the Left/Right path read as a binary number.
const oldUnnamedEntrypointPattern = `@entrypoint\_%`

// UnnamedEntrypoints renames unnamed entrypoints of stored operations from `@entrypoint_K`
// to `entrypoint_N` (leaf ordinal), the name used by `contract.entrypoints`, `GetEntrypointsDocs`
// and `FindByName`. Old names are ambiguous (different leaves of an unbalanced `or` were stored
// under the same K), so every operation's parameters are re-parsed against the parameter type
// of the contract script for the operation's protocol.
type UnnamedEntrypoints struct {
	bulkCount int64
	scripts   map[string]*ast.Script
}

// Key -
func (m *UnnamedEntrypoints) Key() string {
	return "unnamed_entrypoints"
}

// Description -
func (m *UnnamedEntrypoints) Description() string {
	return "rename unnamed entrypoints of operations from @entrypoint_K to entrypoint_N by re-parsing parameters"
}

// Do - migrate function
func (m *UnnamedEntrypoints) Do(ctx *config.Context) error {
	if m.bulkCount == 0 {
		m.bulkCount = 1000
	}
	m.scripts = make(map[string]*ast.Script)

	background := context.Background()

	var (
		lastID        int64
		lastTimestamp time.Time
		updated       int
		skipped       int
	)
	for {
		var ops []operation.Operation
		query := ctx.StorageDB.DB.NewSelect().
			Model(&ops).
			Column("operation.id", "operation.timestamp", "operation.protocol_id", "operation.entrypoint", "operation.parameters").
			Relation("Destination", func(q *bun.SelectQuery) *bun.SelectQuery {
				return q.Column("address")
			}).
			Where("operation.kind = ?", types.OperationKindTransaction).
			Where("operation.entrypoint LIKE ?", oldUnnamedEntrypointPattern)

		if !lastTimestamp.IsZero() {
			// `operation` is a Timescale hypertable partitioned by timestamp:
			// the cursor includes it so the planner can prune chunks.
			query = query.Where("(operation.timestamp, operation.id) > (?, ?)", lastTimestamp, lastID)
		}

		if err := query.
			OrderExpr("operation.timestamp asc, operation.id asc").
			Limit(m.bulkCount).
			Scan(background); err != nil {
			return errors.Wrap(err, "select operations")
		}

		if len(ops) == 0 {
			log.Info().Str("network", ctx.Network.String()).Int("updated", updated).Int("skipped", skipped).
				Msg("unnamed_entrypoints: done")
			return nil
		}

		changed := make([]*operation.Operation, 0, len(ops))
		for i := range ops {
			name, err := m.entrypoint(background, ctx, &ops[i])
			if err != nil {
				skipped++
				log.Warn().Err(err).Str("network", ctx.Network.String()).Int64("operation_id", ops[i].ID).
					Str("entrypoint", ops[i].Entrypoint.String()).Msg("unnamed_entrypoints: can't compute entrypoint, skipping")
				continue
			}
			if err := ops[i].Entrypoint.Set(name); err != nil {
				return err
			}
			changed = append(changed, &ops[i])
		}

		if len(changed) > 0 {
			if err := ctx.StorageDB.DB.RunInTx(background, nil, func(c context.Context, tx bun.Tx) error {
				for _, op := range changed {
					if _, err := tx.NewUpdate().
						Model(op).
						Column("entrypoint").
						WherePK().
						Exec(c); err != nil {
						return errors.Wrapf(err, "update operation id=%d", op.ID)
					}
				}
				return nil
			}); err != nil {
				return err
			}
			updated += len(changed)
		}

		lastID = ops[len(ops)-1].ID
		lastTimestamp = ops[len(ops)-1].Timestamp
		log.Info().Str("network", ctx.Network.String()).Int("updated", updated).Int("skipped", skipped).
			Time("last_timestamp", lastTimestamp).Msg("unnamed_entrypoints: progress")
	}
}

func (m *UnnamedEntrypoints) entrypoint(c context.Context, ctx *config.Context, op *operation.Operation) (string, error) {
	proto, err := ctx.Cache.ProtocolByID(c, op.ProtocolID)
	if err != nil {
		return "", errors.Wrapf(err, "protocol %d", op.ProtocolID)
	}

	key := fmt.Sprintf("%s:%s", op.Destination.Address, proto.SymLink)
	script, ok := m.scripts[key]
	if !ok {
		data, err := ctx.Cache.ScriptBytes(c, op.Destination.Address, proto.SymLink)
		if err != nil {
			return "", errors.Wrapf(err, "script %s", key)
		}
		script, err = ast.NewScriptWithoutCode(data)
		if err != nil {
			return "", errors.Wrapf(err, "parse script %s", key)
		}
		m.scripts[key] = script
	}

	return entrypointByParameters(script, op.Parameters)
}

// entrypointByParameters - the same name the transaction parser stores for applied calls
// (see `Transaction.getEntrypoint` in internal/parsers/operations).
func entrypointByParameters(script *ast.Script, parameters []byte) (string, error) {
	if len(parameters) == 0 {
		return "", errors.New("empty parameters")
	}

	param, err := script.ParameterType()
	if err != nil {
		return "", err
	}

	subTree, err := param.FromParameters(bcdTypes.NewParameters(parameters))
	if err != nil {
		return "", err
	}

	node, name := subTree.UnwrapAndGetEntrypointName()
	if node == nil {
		return "", errors.New("entrypoint is not found")
	}
	return name, nil
}
