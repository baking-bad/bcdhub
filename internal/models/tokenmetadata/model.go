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

// ByName - TokenMetadata sorting filter by Name field
type ByName []TokenMetadata

func (tm ByName) Len() int      { return len(tm) }
func (tm ByName) Swap(i, j int) { tm[i], tm[j] = tm[j], tm[i] }
func (tm ByName) Less(i, j int) bool {
	if tm[i].Name == "" {
		return false
	} else if tm[j].Name == "" {
		return true
	}

	return tm[i].Name < tm[j].Name
}

// ByTokenID - TokenMetadata sorting filter by TokenID field
type ByTokenID []TokenMetadata

func (tm ByTokenID) Len() int           { return len(tm) }
func (tm ByTokenID) Swap(i, j int)      { tm[i], tm[j] = tm[j], tm[i] }
func (tm ByTokenID) Less(i, j int) bool { return tm[i].TokenID < tm[j].TokenID }

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
