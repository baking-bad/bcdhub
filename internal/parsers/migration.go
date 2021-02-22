package parsers

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/contract"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	modelsContract "github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/tidwall/gjson"
)

// MigrationParser -
type MigrationParser struct {
	storage        models.GeneralRepository
	bmdRepo        bigmapdiff.Repository
	filesDirectory string
}

// NewMigrationParser -
func NewMigrationParser(storage models.GeneralRepository, bmdRepo bigmapdiff.Repository, filesDirectory string) *MigrationParser {
	return &MigrationParser{
		storage:        storage,
		bmdRepo:        bmdRepo,
		filesDirectory: filesDirectory,
	}
}

// Parse -
func (p *MigrationParser) Parse(script gjson.Result, old modelsContract.Contract, previous, next protocol.Protocol, timestamp time.Time) ([]models.Model, []models.Model, error) {
	var updates []models.Model
	if previous.SymLink == consts.MetadataAlpha {
		newUpdates, err := p.getUpdates(script, old)
		if err != nil {
			return nil, nil, err
		}
		updates = newUpdates
	}

	newHash, err := contract.ComputeHash([]byte(script.Get("code").Raw))
	if err != nil {
		return nil, nil, err
	}
	if newHash == old.Hash {
		return nil, updates, nil
	}

	migration := migration.Migration{
		ID:          helpers.GenerateID(),
		IndexedTime: time.Now().UnixNano() / 1000,

		Network:      old.Network,
		Level:        previous.EndLevel,
		Protocol:     next.Hash,
		PrevProtocol: previous.Hash,
		Address:      old.Address,
		Timestamp:    timestamp,
		Kind:         consts.MigrationUpdate,
	}

	return []models.Model{&migration}, updates, nil
}

func (p *MigrationParser) getUpdates(script gjson.Result, contract modelsContract.Contract) ([]models.Model, error) {
	storage, err := ast.NewSettledTypedAst(script.Get("code.#(prim==\"storage\").args").Raw, script.Get("storage").Raw)
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
