package block

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Block -
type Block struct {
	Hash        string        `json:"hash"`
	Predecessor string        `json:"predecessor"`
	ChainID     string        `json:"chain_id"`
	Timestamp   time.Time     `json:"timestamp"`
	Network     types.Network `json:"network" gorm:"type:SMALLINT"`
	ID          int64         `json:"-"`
	Level       int64         `json:"level"`
	ProtocolID  int64         `json:"protocol" gorm:"type:SMALLINT"`

	Protocol protocol.Protocol `json:"-"`
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

// ByNetwork - sorting blocks by network. Mainnet - first, others by lexicographical order
type ByNetwork []Block

func (a ByNetwork) Len() int { return len(a) }
func (a ByNetwork) Less(i, j int) bool {
	switch {
	case a[i].Network == a[j].Network:
		return false
	case a[i].Network == types.Mainnet:
		return true
	case a[j].Network == types.Mainnet:
		return false
	default:
		return a[i].Network < a[j].Network
	}
}
func (a ByNetwork) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
