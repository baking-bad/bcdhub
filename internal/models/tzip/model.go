package tzip

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TZIP -
type TZIP struct {
	ID         int64         `json:"-"`
	UpdatedAt  uint64        `gorm:"autoUpdateTime"`
	Level      int64         `json:"level,omitempty"`
	Timestamp  time.Time     `json:"timestamp,omitempty"`
	Address    string        `json:"address" gorm:"index:tzips_contract_idx"`
	Network    types.Network `json:"network" gorm:"type:SMALLINT;index:tzips_contract_idx"`
	Slug       string        `json:"slug,omitempty"`
	DomainName string        `json:"domain_name,omitempty"`
	OffChain   bool          `json:"offchain,omitempty" gorm:",default:false"`
	Extras     types.JSONB   `json:"extras,omitempty" sql:"type:jsonb"`

	TZIP16
	TZIP20
}

// GetID -
func (t *TZIP) GetID() int64 {
	return t.ID
}

// GetIndex -
func (t *TZIP) GetIndex() string {
	return "tzips"
}

// Save -
func (t *TZIP) Save(tx *gorm.DB) error {
	return tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Save(t).Error
}

// LogFields -
func (t *TZIP) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"network": t.Network,
		"address": t.Address,
		"level":   t.Level,
	}
}

// TableName -
func (t *TZIP) TableName() string {
	return "tzips"
}
