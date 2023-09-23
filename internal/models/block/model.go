package block

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/uptrace/bun"
)

// Block -
type Block struct {
	bun.BaseModel `bun:"blocks"`

	ID         int64 `bun:"id,pk,notnull,autoincrement"`
	Hash       string
	Timestamp  time.Time
	Level      int64
	ProtocolID int64 `bun:",type:SMALLINT"`

	Protocol protocol.Protocol `bun:",rel:belongs-to"`
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
func (b *Block) Save(ctx context.Context, tx bun.IDB) error {
	_, err := tx.NewInsert().Model(b).Returning("id").Exec(ctx)
	return err
}

func (Block) PartitionBy() string {
	return ""
}
