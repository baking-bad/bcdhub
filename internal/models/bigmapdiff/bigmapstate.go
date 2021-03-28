package bigmapdiff

import (
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BigMapState -
type BigMapState struct {
	ID              int64       `json:"-" gorm:"autoIncrement:true"`
	Ptr             int64       `json:"ptr" gorm:"not null;primaryKey;autoIncrement:false"`
	LastUpdateLevel int64       `json:"last_update_level" gorm:"last_update_level"`
	Network         string      `json:"network" gorm:"not null;primaryKey"`
	KeyHash         string      `json:"key_hash" gorm:"not null;primaryKey"`
	Contract        string      `json:"contract" gorm:"not null;primaryKey"` // contract is in primary key for supporting alpha protocol (mainnet before babylon)
	Key             types.Bytes `json:"key" gorm:"type:bytes;not null"`
	Value           types.Bytes `json:"value,omitempty" gorm:"type:bytes"`
	Removed         bool        `json:"removed"`

	IsRollback bool `json:"-" gorm:"-"`
}

// GetID -
func (b *BigMapState) GetID() int64 {
	return b.ID
}

// GetIndex -
func (b *BigMapState) GetIndex() string {
	return "big_map_states"
}

// Save -
func (b *BigMapState) Save(tx *gorm.DB) error {
	assign := []string{"last_update_level"}

	switch {
	case b.IsRollback:
		assign = append(assign, "removed", "value")
	case b.Removed:
		assign = append(assign, "removed")
	default:
		assign = append(assign, "value")
	}

	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "network"},
			{Name: "contract"},
			{Name: "ptr"},
			{Name: "key_hash"},
		},
		DoUpdates: clause.AssignmentColumns(assign),
	}).Create(b).Error
}

// GetQueues -
func (b *BigMapState) GetQueues() []string {
	return nil
}

// MarshalToQueue -
func (b *BigMapState) MarshalToQueue() ([]byte, error) {
	return nil, nil
}

// LogFields -
func (b *BigMapState) LogFields() logrus.Fields {
	return logrus.Fields{
		"network":  b.Network,
		"ptr":      b.Ptr,
		"key_hash": b.KeyHash,
		"removed":  b.Removed,
	}
}
