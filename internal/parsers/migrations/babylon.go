package migrations

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/contract"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	modelsContract "github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/go-pg/pg/v10"
)

// Babylon -
type Babylon struct {
	bmdRepo bigmapdiff.Repository
}

// NewBabylon -
func NewBabylon(bmdRepo bigmapdiff.Repository) *Babylon {
	return &Babylon{
		bmdRepo: bmdRepo,
	}
}

// Parse -
func (p *Babylon) Parse(script noderpc.Script, old *modelsContract.Contract, previous, next protocol.Protocol, timestamp time.Time, tx pg.DBI) error {
	if err := p.getUpdates(script, *old, tx); err != nil {
		return err
	}

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

	old.BabylonID = contractScript.ID

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

func (p *Babylon) getUpdates(script noderpc.Script, contract modelsContract.Contract, tx pg.DBI) error {
	storage, err := script.GetSettledStorage()
	if err != nil {
		return err
	}

	ptrs := storage.FindBigMapByPtr()
	if len(ptrs) != 1 {
		return nil
	}
	var newPtr int64
	for p := range ptrs {
		newPtr = p
	}

	bmd, err := p.bmdRepo.GetByAddress(contract.Account.Address)
	if err != nil {
		return err
	}
	if len(bmd) == 0 {
		return nil
	}

	for i := range bmd {
		bmd[i].Ptr = newPtr
		if err := bmd[i].Save(tx); err != nil {
			return err
		}
	}

	keys, err := p.bmdRepo.CurrentByContract(contract.Account.Address)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}

	for i := range keys {
		if _, err := tx.Model(&keys[i]).WherePK().Delete(); err != nil {
			return err
		}

		keys[i].Ptr = newPtr
		if err := keys[i].Save(tx); err != nil {
			return err
		}
	}

	return nil
}
