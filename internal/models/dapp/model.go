package dapp

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DApp -
type DApp struct {
	ID                int64          `json:"-"`
	Name              string         `json:"name"`
	ShortDescription  string         `json:"short_description"`
	FullDescription   string         `json:"full_description"`
	WebSite           string         `json:"web_site"`
	Slug              string         `json:"slug,omitempty"`
	AgoraReviewPostID int64          `json:"agora_review_post_id,omitempty"`
	AgoraQAPostID     int64          `json:"agora_qa_post_id,omitempty"`
	Authors           pq.StringArray `json:"authors" gorm:"type:text[]"`
	SocialLinks       pq.StringArray `json:"social_links" gorm:"type:text[]"`
	Interfaces        pq.StringArray `json:"interfaces" gorm:"type:text[]"`
	Categories        pq.StringArray `json:"categories" gorm:"type:text[]"`
	Contracts         DAppContracts  `json:"contracts" sql:"type:jsonb"`
	Order             int64          `json:"order"`
	Soon              bool           `json:"soon"`

	Pictures  Pictures  `json:"pictures,omitempty" sql:"type:jsonb"`
	DexTokens DexTokens `json:"dex_tokens,omitempty" sql:"type:jsonb"`
}

// GetID -
func (d *DApp) GetID() int64 {
	return d.ID
}

// GetIndex -
func (d *DApp) GetIndex() string {
	return "dapps"
}

// Save -
func (d *DApp) Save(tx *gorm.DB) error {
	return tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Save(d).Error
}

// GetQueues -
func (d *DApp) GetQueues() []string {
	return nil
}

// MarshalToQueue -
func (d *DApp) MarshalToQueue() ([]byte, error) {
	return nil, nil
}

// LogFields -
func (d *DApp) LogFields() logrus.Fields {
	return logrus.Fields{
		"name": d.Name,
	}
}

// Picture -
type Picture struct {
	Link string `json:"link"`
	Type string `json:"type"`
}

// Pictures -
type Pictures []Picture

// Value -
func (j Pictures) Value() (driver.Value, error) {
	if len(j) == 0 {
		return []byte(`[]`), nil
	}
	return json.Marshal(j)
}

// Scan -
func (j *Pictures) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), j)
}

// DexToken -
type DexToken struct {
	TokenID  uint64 `json:"token_id"`
	Contract string `json:"contract"`
}

// DexTokens -
type DexTokens []DexToken

// Value -
func (j DexTokens) Value() (driver.Value, error) {
	if len(j) == 0 {
		return []byte(`[]`), nil
	}
	return json.Marshal(j)
}

// Scan -
func (j *DexTokens) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), j)
}

// DAppContract -
type DAppContract struct {
	Address    string   `json:"address"`
	Entrypoint []string `json:"dex_volume_entrypoints,omitempty"`
	WithTokens bool     `json:"with_tokens,omitempty"`
}

// DAppContracts -
type DAppContracts []DAppContract

// Value -
func (j DAppContracts) Value() (driver.Value, error) {
	if len(j) == 0 {
		return []byte(`[]`), nil
	}
	return json.Marshal(j)
}

// Scan -
func (j *DAppContracts) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), j)
}

// TableName -
func (d *DApp) TableName() string {
	return "dapps"
}
