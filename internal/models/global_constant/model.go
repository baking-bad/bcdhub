package global_constant

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GlobalConstant -
type GlobalConstant struct {
	ID        int64          `json:"-"`
	Network   types.Network  `json:"network" gorm:"type:SMALLINT"`
	Timestamp time.Time      `json:"timestamp"`
	Level     int64          `json:"level"`
	Address   string         `json:"address" gorm:"index:idx_global_constant_address"`
	Value     datatypes.JSON `json:"value,omitempty"`
}

// GetID -
func (m *GlobalConstant) GetID() int64 {
	return m.ID
}

// GetIndex -
func (m *GlobalConstant) GetIndex() string {
	return "global_constants"
}

// Save -
func (m *GlobalConstant) Save(tx *gorm.DB) error {
	return tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Save(m).Error
}

// LogFields -
func (m *GlobalConstant) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"network": m.Network.String(),
		"address": m.Address,
		"block":   m.Level,
	}
}
