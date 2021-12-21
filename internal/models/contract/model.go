package contract

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/global_constant"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/go-pg/pg/v10"
	"github.com/lib/pq"
)

// Contract - entity for contract
type Contract struct {
	// nolint
	tableName struct{} `pg:"contracts"`

	ID        int64
	Network   types.Network `pg:",type:SMALLINT"`
	Level     int64
	Timestamp time.Time

	Hash                 string
	FingerprintCode      []byte
	FingerprintParameter []byte
	FingerprintStorage   []byte
	Tags                 types.Tags     `pg:",use_zero"`
	Entrypoints          pq.StringArray `pg:",type:text[]"`
	FailStrings          pq.StringArray `pg:",type:text[]"`
	Annotations          pq.StringArray `pg:",type:text[]"`
	Hardcoded            pq.StringArray `pg:",type:text[]"`

	Address   string
	Manager   types.NullString `pg:",type:varchar(36)"`
	Delegate  types.NullString `pg:",type:varchar(36)"`
	ProjectID types.NullString `pg:",type:varchar(36)"`

	TxCount         int64 `pg:",use_zero"`
	LastAction      time.Time
	MigrationsCount int64 `pg:",use_zero"`

	Constants []global_constant.GlobalConstant `pg:",many2many:contract_constants"`
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
	_, err := tx.Model(c).
		OnConflict("(id) DO UPDATE").
		Set("project_id = EXCLUDED.project_id").
		Returning("id").
		Insert()
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
