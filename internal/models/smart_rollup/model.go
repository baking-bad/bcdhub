package smartrollup

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/uptrace/bun"
)

// SmartRollup - entity for smart rollup
type SmartRollup struct {
	bun.BaseModel `bun:"smart_rollup"`

	ID        int64 `bun:"id,pk,notnull,autoincrement"`
	Level     int64
	Timestamp time.Time

	Size      uint64
	AddressId int64
	Address   account.Account `bun:",rel:belongs-to"`

	GenesisCommitmentHash string
	PvmKind               string
	Kernel                []byte `bun:",type:bytea"`
	Type                  []byte `bun:",type:bytea"`
}

// GetID -
func (sr *SmartRollup) GetID() int64 {
	return sr.ID
}

// GetIndex -
func (SmartRollup) GetIndex() string {
	return "smart_rollup"
}

// Save -
func (sr *SmartRollup) Save(ctx context.Context, tx bun.IDB) error {
	_, err := tx.NewInsert().Model(sr).On("CONFLICT DO NOTHING").Returning("id").Exec(ctx)
	return err
}

// LogFields -
func (sr *SmartRollup) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"address": sr.Address.Address,
		"block":   sr.Level,
	}
}

func (SmartRollup) PartitionBy() string {
	return ""
}
