package ticket

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/uptrace/bun"
)

type Ticket struct {
	bun.BaseModel `bun:"tickets"`

	ID           int64  `bun:"id,pk,notnull,autoincrement"`
	Level        int64  `bun:"level"`
	TicketerID   int64  `bun:"ticketer_id,unique:ticket_key"`
	ContentType  []byte `bun:"content_type,type:bytea,unique:ticket_key"`
	Content      []byte `bun:"content,type:bytea,unique:ticket_key"`
	UpdatesCount int    `bun:"updates_count"`

	Ticketer account.Account `bun:"rel:belongs-to"`
}

func (t Ticket) GetID() int64 {
	return t.ID
}

func (Ticket) TableName() string {
	return "tickets"
}

func (t Ticket) Hash() string {
	data := make([]byte, len(t.ContentType))
	copy(data, t.ContentType)
	data = append(data, t.Content...)
	data = binary.AppendVarint(data, t.TicketerID)
	h := sha256.New()
	_, _ = h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// LogFields -
func (t Ticket) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"ticketer_id":  t.TicketerID,
		"content":      string(t.Content),
		"content_type": string(t.ContentType),
	}
}
