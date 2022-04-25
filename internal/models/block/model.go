package block

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/go-pg/pg/v10"
)

// Block -
type Block struct {
	// nolint
	tableName struct{} `pg:"blocks"`

	Hash       string
	Timestamp  time.Time
	ID         int64
	Level      int64
	ProtocolID int64 `pg:",type:SMALLINT"`

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

// Save -
func (b *Block) Save(tx pg.DBI) error {
	_, err := tx.Model(b).Returning("id").Insert(b)
	return err
}
