package models

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/sirupsen/logrus"
)

// TZIP -
type TZIP struct {
	Level     int64               `json:"level,omitempty"`
	Timestamp time.Time           `json:"timestamp,omitempty"`
	Address   string              `json:"address"`
	Network   string              `json:"network"`
	Slug      string              `json:"slug,omitempty"`
	Domain    *ReverseTezosDomain `json:"domain,omitempty"`

	tzip.TZIP12
	tzip.TZIP16
	tzip.DAppsTZIP
}

// HasToken -
func (t TZIP) HasToken(network, address string, tokenID int64) bool {
	for i := range t.Tokens.Static {
		if t.Address == address && t.Network == network && t.Tokens.Static[i].TokenID == tokenID {
			return true
		}
	}
	return false
}

// GetID -
func (t *TZIP) GetID() string {
	return fmt.Sprintf("%s_%s", t.Network, t.Address)
}

// GetIndex -
func (t *TZIP) GetIndex() string {
	return "tzip"
}

// GetQueues -
func (t *TZIP) GetQueues() []string {
	return nil
}

// MarshalToQueue -
func (t *TZIP) MarshalToQueue() ([]byte, error) {
	return nil, nil
}

// LogFields -
func (t *TZIP) LogFields() logrus.Fields {
	return logrus.Fields{
		"network": t.Network,
		"address": t.Address,
		"level":   t.Level,
	}
}
