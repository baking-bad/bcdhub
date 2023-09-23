package contract

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

// GlobalConstant -
type GlobalConstant struct {
	bun.BaseModel `bun:"global_constants"`

	ID        int64     `bun:"id,pk,notnull,autoincrement" json:"-"`
	Timestamp time.Time `json:"timestamp"`
	Level     int64     `json:"level"`
	Address   string    `json:"address"`
	Value     []byte    `json:"value,omitempty"`

	Scripts []Script `bun:"m2m:script_constants,join:GlobalConstant=Script"`
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
func (m *GlobalConstant) Save(ctx context.Context, tx bun.IDB) error {
	_, err := tx.NewInsert().Model(m).Returning("id").Exec(ctx)
	return err
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
