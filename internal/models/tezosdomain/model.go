package tezosdomain

import (
	"fmt"
	"time"
)

// TezosDomain -
type TezosDomain struct {
	Name       string            `json:"name"`
	Expiration time.Time         `json:"expiration"`
	Network    string            `json:"network"`
	Address    string            `json:"address"`
	Level      int64             `json:"level"`
	Timestamp  time.Time         `json:"timestamp"`
	Data       map[string]string `json:"data,omitempty"`
}

// GetID -
func (t *TezosDomain) GetID() string {
	return fmt.Sprintf("%s_%s", t.Network, t.Name)
}

// GetIndex -
func (t *TezosDomain) GetIndex() string {
	return "tezos_domain"
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
