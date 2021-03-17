package tezosdomain

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
)

// TezosDomain -
type TezosDomain struct {
	ID         int64       `json:"-"`
	Name       string      `json:"name"`
	Expiration time.Time   `json:"expiration"`
	Network    string      `json:"network"`
	Address    string      `json:"address"`
	Level      int64       `json:"level"`
	Timestamp  time.Time   `json:"timestamp"`
	Data       types.JSONB `json:"data,omitempty" sql:"type:jsonb"`
}

// GetID -
func (t *TezosDomain) GetID() int64 {
	return t.ID
}

// GetIndex -
func (t *TezosDomain) GetIndex() string {
	return "tezos_domains"
}

// GetQueues -
func (t *TezosDomain) GetQueues() []string {
	return nil
}

// MarshalToQueue -
func (t *TezosDomain) MarshalToQueue() ([]byte, error) {
	return nil, nil
}

// ReverseTezosDomain -
type ReverseTezosDomain struct {
	Name       string    `json:"name"`
	Expiration time.Time `json:"expiration"`
}

// DomainsResponse -
type DomainsResponse struct {
	Domains []TezosDomain `json:"domains"`
	Total   int64         `json:"total"`
}
