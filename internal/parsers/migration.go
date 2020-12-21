package parsers

import (
	"encoding/json"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	modelsContract "github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/schema"
	"github.com/baking-bad/bcdhub/internal/parsers/contract"
	"github.com/pkg/errors"
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
	s := schema.Schema{ID: old.Address}
	if err := p.storage.GetByID(&s); err != nil {
		return nil, nil, err
	}

	if err := contract.NewSchemaParser(next.SymLink).Update(script, old.Address, &s); err != nil {
		return nil, nil, err
	}

	var updates []models.Model
	if previous.SymLink == "alpha" {
		newUpdates, err := p.getUpdates(script, old, next, s)
		if err != nil {
			return nil, nil, err
		}
		updates = newUpdates
	}

	newHash, err := contractparser.ComputeContractHash(script.Get("code").Raw)
	if err != nil {
		return nil, nil, err
	}
	if newHash == old.Hash {
		return []models.Model{&s}, updates, nil
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

	return []models.Model{&s, &migration}, updates, nil
}

func (p *MigrationParser) getUpdates(script gjson.Result, contract modelsContract.Contract, protocol protocol.Protocol, s schema.Schema) ([]models.Model, error) {
	stringMetadata, ok := s.Storage[protocol.SymLink]
	if !ok {
		return nil, errors.Errorf("[MigrationParser.getUpdates] Unknown metadata sym link: %s", protocol.SymLink)
	}

	var m meta.Metadata
	if err := json.Unmarshal([]byte(stringMetadata), &m); err != nil {
		return nil, err
	}

	storageJSON := script.Get("storage")
	newMapPtr, err := storage.FindBigMapPointers(m, storageJSON)
	if err != nil {
		return nil, err
	}
	if len(newMapPtr) != 1 {
		return nil, nil
	}
	var newPath string
	var newPtr int64
	for p, b := range newMapPtr {
		newPath = b
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
		bmd[i].BinPath = newPath
		bmd[i].Ptr = newPtr
		updates[i] = &bmd[i]
	}
	return updates, nil
}
