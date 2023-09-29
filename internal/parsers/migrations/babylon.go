package migrations

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/contract"
	"github.com/baking-bad/bcdhub/internal/models"
	modelsContract "github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
)

// Babylon -
type Babylon struct{}

// NewBabylon -
func NewBabylon() *Babylon {
	return &Babylon{}
}

// Parse -
func (p *Babylon) Parse(ctx context.Context, script noderpc.Script, old *modelsContract.Contract, previous, next protocol.Protocol, timestamp time.Time, tx models.Transaction) error {
	if err := p.getUpdates(ctx, script, *old, tx); err != nil {
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
		Level:     next.StartLevel,
		Hash:      newHash,
		Code:      s.Code,
		Storage:   s.Storage,
		Parameter: s.Parameter,
		Views:     s.Views,
	}

	if err := tx.Scripts(ctx, &contractScript); err != nil {
		return err
	}

	old.BabylonID = contractScript.ID

	m := &migration.Migration{
		ContractID:     old.ID,
		Contract:       *old,
		Level:          next.StartLevel,
		ProtocolID:     next.ID,
		PrevProtocolID: previous.ID,
		Timestamp:      timestamp,
		Kind:           types.MigrationKindUpdate,
	}

	return tx.Migrations(ctx, m)
}

// IsMigratable -
func (p *Babylon) IsMigratable(address string) bool {
	return true
}

func (p *Babylon) getUpdates(ctx context.Context, script noderpc.Script, contract modelsContract.Contract, tx models.Transaction) error {
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

	updatedRows, err := tx.BabylonUpdateBigMapDiffs(ctx, contract.Account.Address, newPtr)
	if err != nil {
		return err
	}
	if updatedRows == 0 {
		return nil
	}

	keys, err := tx.DeleteBigMapStatesByContract(ctx, contract.Account.Address)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}

	for i := range keys {
		keys[i].Ptr = newPtr
		keys[i].ID = 0

		if err := tx.BigMapStates(ctx, &keys[i]); err != nil {
			return err
		}
	}

	return nil
}
