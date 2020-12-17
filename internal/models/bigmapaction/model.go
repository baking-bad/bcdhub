package bigmapaction

import (
	"time"
)

// BigMapAction -
type BigMapAction struct {
	ID             string    `json:"-"`
	Action         string    `json:"action"`
	SourcePtr      *int64    `json:"source_ptr,omitempty"`
	DestinationPtr *int64    `json:"destination_ptr,omitempty"`
	OperationID    string    `json:"operation_id"`
	Level          int64     `json:"level"`
	Address        string    `json:"address"`
	Network        string    `json:"network"`
	IndexedTime    int64     `json:"indexed_time"`
	Timestamp      time.Time `json:"timestamp"`
}

// GetID -
func (b *BigMapAction) GetID() string {
	return b.ID
}

// GetIndex -
func (b *BigMapAction) GetIndex() string {
	return "bigmapaction"
}

// GetQueues -
func (b *BigMapAction) GetQueues() []string {
	return nil
}

// MarshalToQueue -
func (b *BigMapAction) MarshalToQueue() ([]byte, error) {
	return nil, nil
}
