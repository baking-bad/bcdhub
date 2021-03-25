package block

import (
	"time"

	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Block -
type Block struct {
	ID int64 `json:"-"`

	Network     string    `json:"network"`
	Hash        string    `json:"hash"`
	Level       int64     `json:"level"`
	Predecessor string    `json:"predecessor"`
	ChainID     string    `json:"chain_id"`
	Protocol    string    `json:"protocol"`
	Timestamp   time.Time `json:"timestamp"`
}

// GetID -
func (b *Block) GetID() int64 {
	return b.ID
}

// GetIndex -
func (b *Block) GetIndex() string {
	return "blocks"
}

// GetQueues -
func (b *Block) GetQueues() []string {
	return []string{"blocks"}
}

// MarshalToQueue -
func (b *Block) MarshalToQueue() ([]byte, error) {
	return json.Marshal(b)
}

// ValidateChainID -
func (b Block) ValidateChainID(chainID string) bool {
	if b.ChainID == "" {
		return b.Level == 0
	}
	return b.ChainID == chainID
}

// Save -
func (b *Block) Save(tx *gorm.DB) error {
	return tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Save(b).Error
}
