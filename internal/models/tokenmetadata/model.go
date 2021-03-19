package tokenmetadata

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/sirupsen/logrus"
)

// TokenMetadata -
type TokenMetadata struct {
	ID                 int64         `json:"-"`
	Network            string        `json:"network"`
	Contract           string        `json:"contract"`
	Level              int64         `json:"level"`
	Timestamp          time.Time     `json:"timestamp"`
	TokenID            uint64        `json:"token_id" gorm:"type:numeric(50,0)"`
	Symbol             string        `json:"symbol"`
	Name               string        `json:"name"`
	Decimals           *int64        `json:"decimals,omitempty"`
	Description        string        `json:"description,omitempty"`
	ArtifactURI        string        `json:"artifact_uri,omitempty"`
	DisplayURI         string        `json:"display_uri,omitempty"`
	ThumbnailURI       string        `json:"thumbnail_uri,omitempty"`
	ExternalURI        string        `json:"external_uri,omitempty"`
	IsTransferable     bool          `json:"is_transferable"`
	IsBooleanAmount    bool          `json:"is_boolean_amount"`
	ShouldPreferSymbol bool          `json:"should_prefer_symbol"`
	Tags               []string      `json:"tags,omitempty"`
	Creators           []string      `json:"creators,omitempty"`
	Formats            []interface{} `json:"formats,omitempty"`
	Extras             types.JSONB   `json:"extras" gorm:"type:jsonb"`
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
func (t *TokenMetadata) GetID() int64 {
	return t.ID
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
