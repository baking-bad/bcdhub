package contract

import (
	"bytes"
	stdJSON "encoding/json"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/go-pg/pg/v10"
)

// Contract - entity for contract
type Contract struct {
	// nolint
	tableName struct{} `pg:"contracts"`

	ID        int64
	Level     int64
	Timestamp time.Time

	AccountID  int64
	Account    account.Account `pg:",rel:has-one"`
	ManagerID  int64
	Manager    account.Account `pg:",rel:has-one"`
	DelegateID int64
	Delegate   account.Account `pg:",rel:has-one"`

	TxCount         int64 `pg:",use_zero"`
	LastAction      time.Time
	MigrationsCount int64      `pg:",use_zero"`
	Tags            types.Tags `pg:",use_zero"`

	AlphaID   int64
	Alpha     Script `pg:",rel:has-one"`
	BabylonID int64
	Babylon   Script `pg:",rel:has-one"`
	JakartaID int64
	Jakarta   Script `pg:",rel:has-one"`
}

// GetID -
func (c *Contract) GetID() int64 {
	return c.ID
}

// GetIndex -
func (c *Contract) GetIndex() string {
	return "contracts"
}

// Save -
func (c *Contract) Save(tx pg.DBI) error {
	_, err := tx.Model(c).OnConflict("DO NOTHING").Returning("id").Insert()
	return err
}

// LogFields -
func (c *Contract) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"address": c.Account,
		"block":   c.Level,
	}
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
