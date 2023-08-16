package smartrollup

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/go-pg/pg/v10"
)

// SmartRollup - entity for smart rollup
type SmartRollup struct {
	// nolint
	tableName struct{} `pg:"smart_rollup"`

	ID        int64
	Level     int64
	Timestamp time.Time

	Size      uint64 `pg:",use_zero"`
	AddressId int64
	Address   account.Account `pg:",rel:has-one"`

	GenesisCommitmentHash string
	PvmKind               string
	Kernel                []byte `pg:",type:bytea"`
	Type                  []byte `pg:",type:bytea"`
}

// GetID -
func (sr *SmartRollup) GetID() int64 {
	return sr.ID
}

// GetIndex -
func (SmartRollup) GetIndex() string {
	return "contracts"
}

// Save -
func (sr *SmartRollup) Save(tx pg.DBI) error {
	_, err := tx.Model(sr).OnConflict("DO NOTHING").Returning("id").Insert()
	return err
}

// LogFields -
func (sr *SmartRollup) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"address": sr.Address.Address,
		"block":   sr.Level,
	}
}
