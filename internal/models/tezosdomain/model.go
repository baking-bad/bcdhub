package tezosdomain

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TezosDomain -
type TezosDomain struct {
	ID         int64         `json:"-" gorm:"autoIncrement:true"`
	Name       string        `json:"name" gorm:"primaryKey"`
	Network    types.Network `json:"network" gorm:"primaryKey;type:SMALLINT"`
	Expiration time.Time     `json:"expiration"`
	Address    string        `json:"address" gorm:"index:tezos_domains_address_idx"`
	Level      int64         `json:"level"`
	Timestamp  time.Time     `json:"timestamp"`
	Data       types.JSONB   `json:"data,omitempty" sql:"type:jsonb"`
}

// GetID -
func (t *TezosDomain) GetID() int64 {
	return t.ID
}

// GetIndex -
func (t *TezosDomain) GetIndex() string {
	return "tezos_domains"
}

// Save -
func (t *TezosDomain) Save(tx *gorm.DB) error {
	var s clause.Set

	if t.Address != "" {
		s = clause.Assignments(map[string]interface{}{
			"address":   t.Address,
			"level":     t.Level,
			"timestamp": t.Timestamp,
			"data":      t.Data,
		})
	} else {
		s = clause.Assignments(map[string]interface{}{
			"expiration": t.Expiration,
		})
	}

	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "network"},
			{Name: "name"},
		},
		DoUpdates: s,
	}).Create(t).Error
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
