package contract

import (
	"time"

	"github.com/uptrace/bun"
)

// GlobalConstant -
type GlobalConstant struct {
	bun.BaseModel `bun:"global_constants"`

	ID        int64     `bun:"id,pk,notnull,autoincrement"`
	Timestamp time.Time `json:"timestamp"`
	Level     int64     `json:"level"`
	Address   string    `bun:"address,type:text"`
	Value     []byte

	Scripts []Script `bun:"m2m:script_constants,join:GlobalConstant=Script"`
}

// GetID -
func (m *GlobalConstant) GetID() int64 {
	return m.ID
}

func (GlobalConstant) TableName() string {
	return "global_constants"
}

// LogFields -
func (m *GlobalConstant) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"address": m.Address,
		"block":   m.Level,
	}
}

func (GlobalConstant) PartitionBy() string {
	return ""
}
