package block

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/go-pg/pg/v10"
)

// Block -
type Block struct {
	// nolint
	tableName struct{} `pg:"blocks"`

	Hash        string
	Predecessor string
	ChainID     string
	Timestamp   time.Time
	Network     types.Network `pg:",type:SMALLINT"`
	ID          int64
	Level       int64
	ProtocolID  int64 `pg:",type:SMALLINT"`

	Protocol protocol.Protocol `pg:",rel:has-one"`
}

// GetID -
func (b *Block) GetID() int64 {
	return b.ID
}

// GetIndex -
func (b *Block) GetIndex() string {
	return "blocks"
}

// ValidateChainID -
func (b Block) ValidateChainID(chainID string) bool {
	if b.ChainID == "" {
		return b.Level == 0
	}
	return b.ChainID == chainID
}

// Save -
func (b *Block) Save(tx pg.DBI) error {
	_, err := tx.Model(b).Returning("id").Insert(b)
	return err
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
