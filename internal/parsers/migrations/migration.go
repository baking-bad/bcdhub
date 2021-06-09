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
	contractParser "github.com/baking-bad/bcdhub/internal/parsers/contract"
	"gorm.io/gorm"
)

// MigrationParser -
type MigrationParser struct {
	storage     models.GeneralRepository
	bmdRepo     bigmapdiff.Repository
	scriptSaver contractParser.ScriptSaver
}

// NewMigrationParser -
func NewMigrationParser(storage models.GeneralRepository, bmdRepo bigmapdiff.Repository, filesDirectory string) *MigrationParser {
	return &MigrationParser{
		storage:     storage,
		bmdRepo:     bmdRepo,
		scriptSaver: contractParser.NewFileScriptSaver(filesDirectory),
	}
}

// Parse -
func (p *MigrationParser) Parse(script noderpc.Script, old modelsContract.Contract, previous, next protocol.Protocol, timestamp time.Time, tx *gorm.DB) error {
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

	if err := p.scriptSaver.Save(codeBytes, contractParser.ScriptSaveContext{
		Hash:    newHash,
		Address: old.Address,
		Network: old.Network.String(),
		SymLink: next.SymLink,
	}); err != nil {
		return err
	}

	if newHash == old.Hash {
		return nil
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

func (p *MigrationParser) getUpdates(script noderpc.Script, contract modelsContract.Contract, tx *gorm.DB) error {
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
		if err := tx.Delete(&keys[i]).Error; err != nil {
			return err
		}

		keys[i].Ptr = newPtr
		if err := keys[i].Save(tx); err != nil {
			return err
		}
	}

	return nil
}
