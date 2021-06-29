package bigmap

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BigMap -
type BigMap struct {
	ID           int64         `json:"-"`
	Network      types.Network `gorm:"uniqueIndex:idx_pointer;type:SMALLINT"`
	Contract     string        `gorm:"uniqueIndex:idx_pointer"`
	Ptr          int64         `gorm:"uniqueIndex:idx_pointer"`
	KeyType      types.Bytes   `gorm:"type:bytes"`
	ValueType    types.Bytes   `gorm:"type:bytes"`
	Tags         types.Tags    `gorm:"default:0"`
	Name         string
	CreatedLevel int64
	CreatedAt    time.Time
}

// GetID -
func (b *BigMap) GetID() int64 {
	return b.ID
}

// GetIndex -
func (b *BigMap) GetIndex() string {
	return "big_maps"
}

// Save -
func (b *BigMap) Save(tx *gorm.DB) error {
	return tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Save(b).Error
}

// GetQueues -
func (b *BigMap) GetQueues() []string {
	return nil
}

// MarshalToQueue -
func (b *BigMap) MarshalToQueue() ([]byte, error) {
	return nil, nil
}

// LogFields -
func (b *BigMap) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"network":  b.Network.String(),
		"contract": b.Contract,
		"ptr":      b.Ptr,
	}
}
