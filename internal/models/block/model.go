package block

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/uptrace/bun"
)

// Block -
type Block struct {
	bun.BaseModel `bun:"blocks"`

	ID         int64     `bun:"id,pk,notnull,autoincrement"`
	Hash       string    `bun:"hash"`
	Timestamp  time.Time `bun:"timestamp,pk,notnull"`
	Level      int64     `bun:"level"`
	ProtocolID int64     `bun:",type:SMALLINT"`

	Protocol protocol.Protocol `bun:",rel:belongs-to"`
}

// GetID -
func (b *Block) GetID() int64 {
	return b.ID
}

func (Block) TableName() string {
	return "blocks"
}
