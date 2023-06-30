package migration

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/go-pg/pg/v10"
)

// Migration -
type Migration struct {
	// nolint
	tableName struct{} `pg:"migrations"`

	ID             int64
	ProtocolID     int64 `pg:",type:SMALLINT"`
	PrevProtocolID int64
	Hash           []byte
	Timestamp      time.Time
	Level          int64
	Kind           types.MigrationKind `pg:",type:SMALLINT"`
	ContractID     int64
	Contract       *contract.Contract `pg:",rel:has-one"`
}

// GetID -
func (m *Migration) GetID() int64 {
	return m.ID
}

// GetIndex -
func (m *Migration) GetIndex() string {
	return "migrations"
}

// Save -
func (m *Migration) Save(tx pg.DBI) error {
	_, err := tx.Model(m).Returning("id").Insert()
	return err
}

// LogFields -
func (m *Migration) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"id":    m.ID,
		"block": m.Level,
		"kind":  m.Kind,
	}
}
