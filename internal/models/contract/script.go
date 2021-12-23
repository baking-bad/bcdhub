package contract

import (
	"bytes"

	"github.com/baking-bad/bcdhub/internal/models/global_constant"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/go-pg/pg/v10"
	"github.com/lib/pq"
)

// Scripts -
type Script struct {
	// nolint
	tableName struct{} `pg:"scripts"`

	ID                   int64
	Hash                 string           `pg:",unique,type:varchar(64)"`
	ProjectID            types.NullString `pg:",type:varchar(36)"`
	Code                 []byte           `pg:",type:bytea"`
	Parameter            []byte           `pg:",type:bytea"`
	Storage              []byte           `pg:",type:bytea"`
	Views                []byte           `pg:",type:bytea"`
	FingerprintCode      []byte           `pg:",type:bytea"`
	FingerprintParameter []byte           `pg:",type:bytea"`
	FingerprintStorage   []byte           `pg:",type:bytea"`
	Entrypoints          pq.StringArray   `pg:",type:text[]"`
	FailStrings          pq.StringArray   `pg:",type:text[]"`
	Annotations          pq.StringArray   `pg:",type:text[]"`
	Hardcoded            pq.StringArray   `pg:",type:text[]"`
	Tags                 types.Tags       `pg:",use_zero"`

	Constants []global_constant.GlobalConstant `pg:",many2many:script_constants"`
}

// GetID -
func (s *Script) GetID() int64 {
	return s.ID
}

// GetIndex -
func (s *Script) GetIndex() string {
	return "scripts"
}

// Save -
func (s *Script) Save(tx pg.DBI) error {
	_, err := tx.Model(s).
		Where("hash = ?hash").
		OnConflict("DO NOTHING").
		Returning("id").SelectOrInsert()
	return err
}

// Full -
func (s *Script) Full() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(`[{"prim":"code","args":`)
	if _, err := buf.Write(s.Code); err != nil {
		return nil, err
	}
	buf.WriteString(`},{"prim":"parameter","args":`)
	if _, err := buf.Write(s.Parameter); err != nil {
		return nil, err
	}
	buf.WriteString(`},{"prim":"storage","args":`)
	if _, err := buf.Write(s.Storage); err != nil {
		return nil, err
	}
	buf.WriteByte('}')
	if len(s.Views) > 2 {
		if _, err := buf.Write(s.Views[1 : len(s.Views)-1]); err != nil {
			return nil, err
		}
	}
	buf.WriteByte(']')

	return buf.Bytes(), nil
}
