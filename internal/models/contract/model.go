package contract

import (
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
