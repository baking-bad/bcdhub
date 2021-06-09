package migration

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/sirupsen/logrus"
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

// GetQueues -
func (m *Migration) GetQueues() []string {
	return []string{"migrations"}
}

// MarshalToQueue -
func (m *Migration) MarshalToQueue() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", m.ID)), nil
}

// LogFields -
func (m *Migration) LogFields() logrus.Fields {
	return logrus.Fields{
		"network": m.Network.String(),
		"address": m.Address,
		"block":   m.Level,
		"kind":    m.Kind,
	}
}
