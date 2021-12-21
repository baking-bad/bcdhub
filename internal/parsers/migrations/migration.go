package migrations

import (
	"encoding/json"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/contract"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	modelsContract "github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
)

// MigrationParser -
type MigrationParser struct {
	storage models.GeneralRepository
	bmdRepo bigmapdiff.Repository
}

// NewMigrationParser -
func NewMigrationParser(storage models.GeneralRepository, bmdRepo bigmapdiff.Repository) *MigrationParser {
	return &MigrationParser{
		storage: storage,
		bmdRepo: bmdRepo,
	}
}

// Parse -
func (p *MigrationParser) Parse(script noderpc.Script, old modelsContract.Contract, previous, next protocol.Protocol, timestamp time.Time, tx pg.DBI) error {
	if previous.SymLink == bcd.SymLinkAlpha {
		if err := p.getUpdates(script, old, tx); err != nil {
			return err
		}
	}

	codeBytes, err := json.Marshal(script.Code)
	if err != nil {
		return err
	}

	newHash, err := contract.ComputeHash(codeBytes)
	if err != nil {
		return err
	}

	contractScript := modelsContract.Script{
		Hash: newHash,
		Code: codeBytes,
	}

	if err := contractScript.Save(tx); err != nil {
		return err
	}

	switch next.SymLink {
	case bcd.SymLinkAlpha:
		if contractScript.ID == old.AlphaID {
			return nil
		}
	case bcd.SymLinkBabylon:
		if contractScript.ID == old.BabylonID {
			return nil
		}
	default:
		return errors.Errorf("unknown protocol symbolic link: %s", next.SymLink)
	}

	m := &migration.Migration{
		Network:        old.Network,
		Level:          previous.EndLevel,
		ProtocolID:     next.ID,
		PrevProtocolID: previous.ID,
		Address:        old.Address,
		Timestamp:      timestamp,
		Kind:           types.MigrationKindUpdate,
	}

	return m.Save(tx)
}

func (p *MigrationParser) getUpdates(script noderpc.Script, contract modelsContract.Contract, tx pg.DBI) error {
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

	bmd, err := p.bmdRepo.GetByAddress(contract.Network, contract.Address)
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

	keys, err := p.bmdRepo.CurrentByContract(contract.Network, contract.Address)
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
