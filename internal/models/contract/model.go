package contract

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// Contract - entity for contract
type Contract struct {
	Network   string    `json:"network"`
	Level     int64     `json:"level"`
	Timestamp time.Time `json:"timestamp"`
	Language  string    `json:"language,omitempty"`

	Hash        string       `json:"hash"`
	Fingerprint *Fingerprint `json:"fingerprint,omitempty"`
	Tags        []string     `json:"tags,omitempty"`
	Hardcoded   []string     `json:"hardcoded,omitempty"`
	FailStrings []string     `json:"fail_strings,omitempty"`
	Annotations []string     `json:"annotations,omitempty"`
	Entrypoints []string     `json:"entrypoints,omitempty"`

	Address  string `json:"address"`
	Manager  string `json:"manager,omitempty"`
	Delegate string `json:"delegate,omitempty"`

	ProjectID          string    `json:"project_id,omitempty"`
	TxCount            int64     `json:"tx_count"`
	LastAction         time.Time `json:"last_action"`
	FoundBy            string    `json:"found_by,omitempty"`
	MigrationsCount    int64     `json:"migrations_count,omitempty"`
	Alias              string    `json:"alias,omitempty"`
	DelegateAlias      string    `json:"delegate_alias,omitempty"`
	Verified           bool      `json:"verified,omitempty"`
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
func (c *Contract) GetID() string {
	return fmt.Sprintf("%s_%s", c.Network, c.Address)
}

// GetIndex -
func (c *Contract) GetIndex() string {
	return "contract"
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
	return []byte(c.GetID()), nil
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
