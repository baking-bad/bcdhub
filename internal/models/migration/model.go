package migration

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/uptrace/bun"
)

// Migration -
type Migration struct {
	bun.BaseModel `bun:"migrations"`

	ID             int64 `bun:"id,pk,notnull,autoincrement"`
	ProtocolID     int64 `bun:"protocol_id,type:SMALLINT"`
	PrevProtocolID int64
	Hash           []byte
	Timestamp      time.Time `bun:"timestamp,pk,notnull"`
	Level          int64
	Kind           types.MigrationKind `bun:"kind,type:SMALLINT"`
	ContractID     int64
	Contract       *contract.Contract `bun:"rel:belongs-to"`
}

// GetID -
func (m *Migration) GetID() int64 {
	return m.ID
}

// TableName -
func (Migration) TableName() string {
	return "migrations"
}

// LogFields -
func (m *Migration) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"id":    m.ID,
		"block": m.Level,
		"kind":  m.Kind,
	}
}
