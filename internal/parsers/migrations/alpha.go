package migrations

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/contract"
	modelsContract "github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/go-pg/pg/v10"
)

// Alpha -
type Alpha struct{}

// NewAlpha -
func NewAlpha() *Alpha {
	return &Alpha{}
}

// Parse -
func (p *Alpha) Parse(script noderpc.Script, old *modelsContract.Contract, previous, next protocol.Protocol, timestamp time.Time, tx pg.DBI) error {
	codeBytes, err := json.Marshal(script.Code)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := json.Compact(&buf, codeBytes); err != nil {
		return err
	}

	newHash, err := contract.ComputeHash(buf.Bytes())
	if err != nil {
		return err
	}

	var s bcd.RawScript
	if err := json.Unmarshal(buf.Bytes(), &s); err != nil {
		return err
	}

	contractScript := modelsContract.Script{
		Hash:      newHash,
		Code:      s.Code,
		Storage:   s.Storage,
		Parameter: s.Parameter,
		Views:     s.Views,
	}

	if err := contractScript.Save(tx); err != nil {
		return err
	}

	old.AlphaID = contractScript.ID

	m := &migration.Migration{
		ContractID:     old.ID,
		Level:          previous.EndLevel,
		ProtocolID:     next.ID,
		PrevProtocolID: previous.ID,
		Timestamp:      timestamp,
		Kind:           types.MigrationKindUpdate,
	}

	return m.Save(tx)
}

// IsMigratable -
func (p *Alpha) IsMigratable(address string) bool {
	return true
}
