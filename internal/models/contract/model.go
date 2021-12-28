package contract

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/go-pg/pg/v10"
)

// Contract - entity for contract
type Contract struct {
	// nolint
	tableName struct{} `pg:"contracts"`

	ID        int64
	Network   types.Network `pg:",type:SMALLINT"`
	Level     int64
	Timestamp time.Time

	Address  string
	Manager  types.NullString `pg:",type:varchar(36)"`
	Delegate types.NullString `pg:",type:varchar(36)"`

	TxCount         int64 `pg:",use_zero"`
	LastAction      time.Time
	MigrationsCount int64      `pg:",use_zero"`
	Tags            types.Tags `pg:",use_zero"`

	AlphaID   int64
	Alpha     Script `pg:",rel:has-one"`
	BabylonID int64
	Babylon   Script `pg:",rel:has-one"`
}

// NewEmptyContract -
func NewEmptyContract(network types.Network, address string) Contract {
	return Contract{
		Network: network,
		Address: address,
	}
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
	_, err := tx.Model(c).Returning("id").Insert()
	return err
}

// LogFields -
func (c *Contract) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"network": c.Network.String(),
		"address": c.Address,
		"block":   c.Level,
	}
}

// Fingerprint -
type Fingerprint struct {
	Code      string
	Storage   string
	Parameter string
}

// Compare -
func (f *Fingerprint) Compare(second *Fingerprint) bool {
	return f.Code == second.Code && f.Parameter == second.Parameter && f.Storage == second.Storage
}
