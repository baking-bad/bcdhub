package bigmap

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Action -
type Action struct {
	ID            int64              `json:"-"`
	Action        types.BigMapAction `json:"action" gorm:"type:SMALLINT"`
	SourceID      *int64             `json:"-" gorm:"not null;index:bm_action_src_id_key_idx"`
	DestinationID *int64             `json:"-"`
	OperationID   int64              `json:"operation_id"`
	Level         int64              `json:"level"`
	Timestamp     time.Time          `json:"timestamp"`

	Source      BigMap
	Destination BigMap
}

// TableName -
func (Action) TableName() string {
	return "big_map_actions"
}

// GetID -
func (b *Action) GetID() int64 {
	return b.ID
}

// GetIndex -
func (b *Action) GetIndex() string {
	return "big_map_actions"
}

// Save -
func (b *Action) Save(tx *gorm.DB) error {
	return tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Save(b).Error
}
