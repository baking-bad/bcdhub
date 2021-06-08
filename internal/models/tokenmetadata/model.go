package tokenmetadata

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TokenMetadata -
type TokenMetadata struct {
	ID                 int64          `json:"-" gorm:"autoIncrement:true;not null;"`
	Network            types.Network  `json:"network" gorm:"type:SMALLINT;primaryKey;default:0"`
	Contract           string         `json:"contract" gorm:"primaryKey"`
	TokenID            uint64         `json:"token_id" gorm:"type:numeric(50,0);primaryKey"`
	Level              int64          `json:"level"`
	Timestamp          time.Time      `json:"timestamp"`
	Symbol             string         `json:"symbol"`
	Name               string         `json:"name"`
	Decimals           *int64         `json:"decimals,omitempty"`
	Description        string         `json:"description,omitempty"`
	ArtifactURI        string         `json:"artifact_uri,omitempty"`
	DisplayURI         string         `json:"display_uri,omitempty"`
	ThumbnailURI       string         `json:"thumbnail_uri,omitempty"`
	ExternalURI        string         `json:"external_uri,omitempty"`
	IsTransferable     bool           `json:"is_transferable" gorm:"default:true"`
	IsBooleanAmount    bool           `json:"is_boolean_amount"`
	ShouldPreferSymbol bool           `json:"should_prefer_symbol"`
	Tags               pq.StringArray `json:"tags,omitempty" gorm:"type:text[]"`
	Creators           pq.StringArray `json:"creators,omitempty" gorm:"type:text[]"`
	Formats            types.Bytes    `json:"formats,omitempty" gorm:"type:bytes"`
	Extras             types.JSONB    `json:"extras" gorm:"type:jsonb"`
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

// Save -
func (t *TokenMetadata) Save(tx *gorm.DB) error {
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "network"},
			{Name: "contract"},
			{Name: "token_id"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"symbol", "name", "decimals", "description", "artifact_uri", "display_uri", "thumbnail_uri", "external_uri",
			"is_transferable", "is_boolean_amount", "should_prefer_symbol", "tags", "creators", "formats", "extras",
		}),
	}).Save(t).Error
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
		"network":  t.Network.String(),
		"contract": t.Contract,
		"token_id": t.TokenID,
	}
}
