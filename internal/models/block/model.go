package block

import (
	"time"
)

// Block -
type Block struct {
	ID string `json:"-"`

	Network     string    `json:"network"`
	Hash        string    `json:"hash"`
	Level       int64     `json:"level"`
	Predecessor string    `json:"predecessor"`
	ChainID     string    `json:"chain_id"`
	Protocol    string    `json:"protocol"`
	Timestamp   time.Time `json:"timestamp"`
}

// GetID -
func (b *Block) GetID() string {
	return b.ID
}

// GetIndex -
func (b *Block) GetIndex() string {
	return "block"
}

// GetQueues -
func (b *Block) GetQueues() []string {
	return nil
}

// MarshalToQueue -
func (b *Block) MarshalToQueue() ([]byte, error) {
	return nil, nil
}

// ValidateChainID -
func (b Block) ValidateChainID(chainID string) bool {
	if b.ChainID == "" {
		return b.Level == 0
	}
	return b.ChainID == chainID
}
