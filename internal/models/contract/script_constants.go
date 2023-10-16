package contract

import (
	"context"

	"github.com/uptrace/bun"
)

// ScriptConstants -
type ScriptConstants struct {
	bun.BaseModel `bun:"script_constants"`

	ScriptId         int64
	Script           Script `bun:"rel:belongs-to,join:script_id=id"`
	GlobalConstantId int64
	GlobalConstant   GlobalConstant `bun:"rel:belongs-to,join:global_constant_id=id"`
}

// GetID -
func (ScriptConstants) GetID() int64 {
	return 0
}

func (ScriptConstants) TableName() string {
	return "script_constants"
}

// Save -
func (ScriptConstants) Save(ctx context.Context, tx bun.IDB) error {
	return nil
}
