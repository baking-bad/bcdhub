package parsers

import (
	"encoding/json"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/contract"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	modelsContract "github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	contractParser "github.com/baking-bad/bcdhub/internal/parsers/contract"
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
func (p *MigrationParser) Parse(script noderpc.Script, old modelsContract.Contract, previous, next protocol.Protocol, timestamp time.Time) ([]models.Model, error) {
	updates := make([]models.Model, 0)

	if previous.SymLink == consts.MetadataAlpha {
		newUpdates, err := p.getUpdates(script, old)
		if err != nil {
			return nil, err
		}
		updates = append(updates, newUpdates...)
	}

	codeBytes, err := json.Marshal(script.Code)
	if err != nil {
		return nil, err
	}

	newHash, err := contract.ComputeHash(codeBytes)
	if err != nil {
		return nil, err
	}

	if err := p.scriptSaver.Save(codeBytes, contractParser.ScriptSaveContext{
		Hash:    newHash,
		Address: old.Address,
		Network: old.Network,
		SymLink: next.SymLink,
	}); err != nil {
		return nil, err
	}

	if newHash == old.Hash {
		return updates, nil
	}

	updates = append(updates, &migration.Migration{
		Network:      old.Network,
		Level:        previous.EndLevel,
		Protocol:     next.Hash,
		PrevProtocol: previous.Hash,
		Address:      old.Address,
		Timestamp:    timestamp,
		Kind:         consts.MigrationUpdate,
	})

	return updates, nil
}

func (p *MigrationParser) getUpdates(script noderpc.Script, contract modelsContract.Contract) ([]models.Model, error) {
	storage, err := script.GetSettledStorage()
	if err != nil {
		return nil, err
	}

	ptrs := storage.FindBigMapByPtr()
	if len(ptrs) != 1 {
		return nil, nil
	}
	var newPtr int64
	for p := range ptrs {
		newPtr = p
	}

	bmd, err := p.bmdRepo.GetByAddress(contract.Network, contract.Address)
	if err != nil {
		return nil, err
	}
	if len(bmd) == 0 {
		return nil, nil
	}

	updates := make([]models.Model, len(bmd))
	for i := range bmd {
		bmd[i].Ptr = newPtr
		updates[i] = &bmd[i]
	}
	return updates, nil
}
