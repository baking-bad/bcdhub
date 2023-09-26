package contract

import (
	"bytes"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/lib/pq"
	"github.com/uptrace/bun"
)

// Scripts -
type Script struct {
	bun.BaseModel `bun:"scripts"`

	ID          int64          `bun:"id,pk,notnull,autoincrement"`
	Level       int64          `bun:"level"`
	Hash        string         `bun:"hash"`
	Code        []byte         `bun:",type:bytea"`
	Parameter   []byte         `bun:",type:bytea"`
	Storage     []byte         `bun:",type:bytea"`
	Views       []byte         `bun:",type:bytea"`
	Entrypoints pq.StringArray `bun:",type:text[]"`
	FailStrings pq.StringArray `bun:",type:text[]"`
	Annotations pq.StringArray `bun:",type:text[]"`
	Hardcoded   pq.StringArray `bun:",type:text[]"`
	Tags        types.Tags

	Constants []GlobalConstant `bun:"m2m:script_constants,join:Script=GlobalConstant"`
}

// GetID -
func (s *Script) GetID() int64 {
	return s.ID
}

func (Script) TableName() string {
	return "scripts"
}

// Full -
func (s *Script) Full() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(`[{"prim":"parameter","args":`)
	if _, err := buf.Write(s.Parameter); err != nil {
		return nil, err
	}
	buf.WriteString(`},{"prim":"storage","args":`)
	if _, err := buf.Write(s.Storage); err != nil {
		return nil, err
	}
	buf.WriteString(`},{"prim":"code","args":`)
	if _, err := buf.Write(s.Code); err != nil {
		return nil, err
	}
	buf.WriteByte('}')
	if len(s.Views) > 2 {
		buf.WriteByte(',')
		if _, err := buf.Write(s.Views[1 : len(s.Views)-1]); err != nil {
			return nil, err
		}
	}
	buf.WriteByte(']')

	return buf.Bytes(), nil
}
