package account

import (
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/uptrace/bun"
)

// Account -
type Account struct {
	bun.BaseModel `bun:"accounts"`

	ID      int64             `bun:"id,pk,notnull,autoincrement"`
	Type    types.AccountType `bun:"type,type:SMALLINT"`
	Address string            `bun:"address"`
	Alias   string
}

// GetID -
func (a *Account) GetID() int64 {
	return a.ID
}

// GetIndex -
func (a *Account) GetIndex() string {
	return "accounts"
}

// IsEmpty -
func (a *Account) IsEmpty() bool {
	return a.Address == "" || a.Type == types.AccountTypeUnknown
}

func (Account) PartitionBy() string {
	return ""
}
