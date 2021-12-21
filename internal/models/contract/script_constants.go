package contract

import "github.com/go-pg/pg/v10"

// ScriptConstants -
type ScriptConstants struct {
	// nolint
	tableName struct{} `pg:"script_constants"`

	ScriptId         int64
	GlobalConstantId int64
}

// GetID -
func (ScriptConstants) GetID() int64 {
	return 0
}

// GetIndex -
func (ScriptConstants) GetIndex() string {
	return "script_constants"
}

// Save -
func (ScriptConstants) Save(tx pg.DBI) error {
	return nil
}
