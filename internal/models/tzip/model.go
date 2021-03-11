package tzip

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
	"github.com/sirupsen/logrus"
)

// TZIP -
type TZIP struct {
	Level     int64                           `json:"level,omitempty"`
	Timestamp time.Time                       `json:"timestamp,omitempty"`
	Address   string                          `json:"address"`
	Network   string                          `json:"network"`
	Slug      string                          `json:"slug,omitempty"`
	Domain    *tezosdomain.ReverseTezosDomain `json:"domain,omitempty"`
	OffChain  bool                            `json:"offchain,omitempty"`
	Extras    map[string]interface{}          `json:"extras,omitempty"`

	TZIP16
	TZIP20
	DAppsTZIP
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
