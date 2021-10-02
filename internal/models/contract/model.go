package contract

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Contract - entity for contract
type Contract struct {
	ID        int64         `json:"-"`
	Network   types.Network `json:"network" gorm:"type:SMALLINT;index:contracts_idx;index:idx_contracts_level_network"`
	Level     int64         `json:"level" gorm:"index:idx_contracts_level_network"`
	Timestamp time.Time     `json:"timestamp"`
	Language  string        `json:"language,omitempty"`

	Hash                 string         `json:"hash"`
	FingerprintCode      []byte         `json:"fgpt_code,omitempty"`
	FingerprintParameter []byte         `json:"fgpt_parameter,omitempty"`
	FingerprintStorage   []byte         `json:"fgpt_storage,omitempty"`
	Tags                 types.Tags     `json:"tags,omitempty" gorm:"default:0"`
	Entrypoints          pq.StringArray `json:"entrypoints,omitempty" gorm:"type:text[]"`
	FailStrings          pq.StringArray `json:"fail_strings,omitempty" gorm:"type:text[]"`
	Annotations          pq.StringArray `json:"annotations,omitempty" gorm:"type:text[]"`
	Hardcoded            pq.StringArray `json:"hardcoded,omitempty" gorm:"type:text[]"`

	Address  string `json:"address" gorm:"index:contracts_idx"`
	Manager  string `json:"manager,omitempty"`
	Delegate string `json:"delegate,omitempty"`

	ProjectID       string    `json:"project_id,omitempty"`
	TxCount         int64     `json:"tx_count" gorm:",default:0"`
	LastAction      time.Time `json:"last_action"`
	MigrationsCount int64     `json:"migrations_count,omitempty" gorm:",default:0"`
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
func (t *Contract) Save(tx *gorm.DB) error {
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"project_id", "alias", "delegate_alias", "verified", "verification_source"}),
	}).Save(t).Error
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
	Code      string `json:"code"`
	Storage   string `json:"storage"`
	Parameter string `json:"parameter"`
}

// Compare -
func (f *Fingerprint) Compare(second *Fingerprint) bool {
	return f.Code == second.Code && f.Parameter == second.Parameter && f.Storage == second.Storage
}

// Light -
type Light struct {
	Address  string        `json:"address"`
	Network  types.Network `json:"network"`
	Deployed time.Time     `json:"deploy_time"`
}
