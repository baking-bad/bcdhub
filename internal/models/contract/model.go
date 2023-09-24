package contract

import (
	"bytes"
	stdJSON "encoding/json"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/uptrace/bun"
)

// Contract - entity for contract
type Contract struct {
	bun.BaseModel `bun:"contracts"`

	ID        int64 `bun:"id,pk,notnull,autoincrement"`
	Level     int64
	Timestamp time.Time

	AccountID  int64
	Account    account.Account `bun:"rel:belongs-to"`
	ManagerID  int64
	Manager    account.Account `bun:"rel:belongs-to"`
	DelegateID int64
	Delegate   account.Account `bun:"rel:belongs-to"`

	TxCount         int64
	LastAction      time.Time
	MigrationsCount int64
	Tags            types.Tags

	AlphaID   int64
	Alpha     Script `bun:"rel:belongs-to"`
	BabylonID int64
	Babylon   Script `bun:"rel:belongs-to"`
	JakartaID int64
	Jakarta   Script `bun:"rel:belongs-to"`
}

// GetID -
func (c *Contract) GetID() int64 {
	return c.ID
}

// GetIndex -
func (c *Contract) GetIndex() string {
	return "contracts"
}

// LogFields -
func (c *Contract) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"address": c.Account,
		"block":   c.Level,
	}
}

func (Contract) PartitionBy() string {
	return ""
}

// CurrentScript -
func (c *Contract) CurrentScript() *Script {
	switch {
	case c.JakartaID > 0:
		return &c.Jakarta
	case c.BabylonID > 0:
		return &c.Babylon
	case c.AlphaID > 0:
		return &c.Alpha
	default:
		return nil
	}
}

// Sections -
type Sections struct {
	Parameter  stdJSON.RawMessage `json:"parameter"`
	ReturnType stdJSON.RawMessage `json:"returnType"`
	Code       stdJSON.RawMessage `json:"code"`
}

var null = []byte("null")

// Empty -
func (s Sections) Empty() bool {
	return bytes.HasSuffix(s.Code, null) && bytes.HasSuffix(s.Parameter, null) && bytes.HasSuffix(s.ReturnType, null)
}

// IsParameterEmpty -
func (s Sections) IsParameterEmpty() bool {
	return s.Parameter == nil || bytes.HasSuffix(s.Parameter, null)
}

// Views -
type Views []View

// View -
type View struct {
	Name            string               `json:"name"`
	Description     string               `json:"description"`
	Implementations []ViewImplementation `json:"implementations"`
}

// ViewImplementation -
type ViewImplementation struct {
	MichelsonStorageView Sections `json:"michelsonStorageView"`
}

type Update struct {
	bun.BaseModel `bun:"contracts"`

	AccountID  int64
	LastAction time.Time
	TxCount    uint64
}
