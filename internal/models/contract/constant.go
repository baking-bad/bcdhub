package contract

import (
	"time"

	"github.com/go-pg/pg/v10"
)

// GlobalConstant -
type GlobalConstant struct {
	// nolint
	tableName struct{} `pg:"global_constants"`

	ID        int64     `json:"-"`
	Timestamp time.Time `json:"timestamp"`
	Level     int64     `json:"level"`
	Address   string    `json:"address"`
	Value     []byte    `json:"value,omitempty"`

	Scripts []Script `pg:",many2many:script_constants"`
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
func (m *GlobalConstant) Save(tx pg.DBI) error {
	_, err := tx.Model(m).Returning("id").Insert()
	return err
}

// LogFields -
func (m *GlobalConstant) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"address": m.Address,
		"block":   m.Level,
	}
}
