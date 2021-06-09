package bigmapaction

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BigMapAction -
type BigMapAction struct {
	ID             int64         `json:"-"`
	Action         string        `json:"action"`
	SourcePtr      *int64        `json:"source_ptr,omitempty"`
	DestinationPtr *int64        `json:"destination_ptr,omitempty"`
	OperationID    int64         `json:"operation_id"`
	Level          int64         `json:"level"`
	Address        string        `json:"address"`
	Network        types.Network `json:"network" gorm:"type:SMALLINT"`
	Timestamp      time.Time     `json:"timestamp"`
}

// GetID -
func (b *BigMapAction) GetID() int64 {
	return b.ID
}

// GetIndex -
func (b *BigMapAction) GetIndex() string {
	return "big_map_actions"
}

// GetQueues -
func (b *BigMapAction) GetQueues() []string {
	return nil
}

// MarshalToQueue -
func (b *BigMapAction) MarshalToQueue() ([]byte, error) {
	return nil, nil
}

// Save -
func (b *BigMapAction) Save(tx *gorm.DB) error {
	return tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Save(b).Error
}
