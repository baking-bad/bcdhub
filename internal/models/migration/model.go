package migration

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Migration -
type Migration struct {
	ID             int64               `json:"-"`
	Network        types.Network       `json:"network" gorm:"type:SMALLINT"`
	ProtocolID     int64               `json:"protocol" gorm:"type:SMALLINT"`
	PrevProtocolID int64               `json:"prev_protocol,omitempty"`
	Hash           string              `json:"hash,omitempty"`
	Timestamp      time.Time           `json:"timestamp"`
	Level          int64               `json:"level"`
	Address        string              `json:"address"`
	Kind           types.MigrationKind `json:"kind" gorm:"type:SMALLINT"`
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
func (m *Migration) Save(tx *gorm.DB) error {
	return tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Save(m).Error
}

// LogFields -
func (m *Migration) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"network": m.Network.String(),
		"address": m.Address,
		"block":   m.Level,
		"kind":    m.Kind,
	}
}
