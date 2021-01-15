package tokenmetadata

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// TokenMetadata -
type TokenMetadata struct {
	Network   string                 `json:"network"`
	Contract  string                 `json:"contract"`
	Level     int64                  `json:"level"`
	Timestamp time.Time              `json:"timestamp"`
	TokenID   int64                  `json:"token_id"`
	Symbol    string                 `json:"symbol"`
	Name      string                 `json:"name"`
	Decimals  *int64                 `json:"decimals,omitempty"`
	Extras    map[string]interface{} `json:"extras"`
}

// GetID -
func (t *TokenMetadata) GetID() string {
	return fmt.Sprintf("%s_%s_%d", t.Network, t.Contract, t.TokenID)
}

// GetIndex -
func (t *TokenMetadata) GetIndex() string {
	return "token_metadata"
}

// GetQueues -
func (t *TokenMetadata) GetQueues() []string {
	return nil
}

// MarshalToQueue -
func (t *TokenMetadata) MarshalToQueue() ([]byte, error) {
	return nil, nil
}

// LogFields -
func (t *TokenMetadata) LogFields() logrus.Fields {
	return logrus.Fields{
		"network":  t.Network,
		"contract": t.Contract,
		"token_id": t.TokenID,
	}
}
