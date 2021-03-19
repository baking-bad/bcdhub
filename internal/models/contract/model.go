package contract

import (
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// Contract - entity for contract
type Contract struct {
	ID        int64     `json:"-"`
	Network   string    `json:"network"`
	Level     int64     `json:"level"`
	Timestamp time.Time `json:"timestamp"`
	Language  string    `json:"language,omitempty"`

	Hash                 string         `json:"hash"`
	FingerprintCode      []byte         `json:"fgpt_code,omitempty"`
	FingerprintParameter []byte         `json:"fgpt_parameter,omitempty"`
	FingerprintStorage   []byte         `json:"fgpt_storage,omitempty"`
	Tags                 pq.StringArray `json:"tags,omitempty" gorm:"type:text[]"`
	Entrypoints          pq.StringArray `json:"entrypoints,omitempty" gorm:"type:text[]"`
	FailStrings          pq.StringArray `json:"fail_strings,omitempty" gorm:"type:text[]"`
	Annotations          pq.StringArray `json:"annotations,omitempty" gorm:"type:text[]"`

	Address  string `json:"address"`
	Manager  string `json:"manager,omitempty"`
	Delegate string `json:"delegate,omitempty"`

	ProjectID          string    `json:"project_id,omitempty"`
	TxCount            int64     `json:"tx_count" gorm:",default:0"`
	LastAction         time.Time `json:"last_action"`
	MigrationsCount    int64     `json:"migrations_count,omitempty" gorm:",default:0"`
	Alias              string    `json:"alias,omitempty"`
	DelegateAlias      string    `json:"delegate_alias,omitempty"`
	Verified           bool      `json:"verified,omitempty" gorm:",default:false"`
	VerificationSource string    `json:"verification_source,omitempty"`
}

// NewEmptyContract -
func NewEmptyContract(network, address string) Contract {
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

// GetQueues -
func (c *Contract) GetQueues() []string {
	return []string{"contracts", "projects"}
}

// LogFields -
func (c *Contract) LogFields() logrus.Fields {
	return logrus.Fields{
		"network": c.Network,
		"address": c.Address,
		"block":   c.Level,
	}
}

// MarshalToQueue -
func (c *Contract) MarshalToQueue() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", c.ID)), nil
}

// IsFA12 - checks contract realizes fa12 interface
func (c *Contract) IsFA12() bool {
	for i := range c.Tags {
		if c.Tags[i] == "fa12" {
			return true
		}
	}
	return false
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
	Address  string    `json:"address"`
	Network  string    `json:"network"`
	Deployed time.Time `json:"deploy_time"`
}
