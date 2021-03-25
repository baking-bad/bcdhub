package bigmapdiff

import (
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BigMapState -
type BigMapState struct {
	ID       int64       `json:"-" gorm:"autoIncrement:true"`
	Ptr      int64       `json:"ptr" gorm:"not null;primaryKey;autoIncrement:false"`
	Network  string      `json:"network" gorm:"not null;primaryKey"`
	KeyHash  string      `json:"key_hash" gorm:"not null;primaryKey"`
	Contract string      `json:"contract"`
	Key      types.Bytes `json:"key" gorm:"type:bytes;not null"`
	Value    types.Bytes `json:"value,omitempty" gorm:"type:bytes"`
	Removed  bool        `json:"removed"`
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
	return tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Save(b).Error
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
