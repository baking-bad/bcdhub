package tokens

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/schema"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	tzipStorage "github.com/baking-bad/bcdhub/internal/parsers/tzip/storage"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// Parser -
type Parser struct {
	bmdRepo      bigmapdiff.Repository
	blocksRepo   block.Repository
	protocolRepo protocol.Repository
	schemaRepo   schema.Repository
	storage      models.GeneralRepository

	rpc       noderpc.INode
	sharePath string
	network   string
	ipfs      []string
}

// NewParser -
func NewParser(bmdRepo bigmapdiff.Repository, blocksRepo block.Repository, protocolRepo protocol.Repository, schemaRepo schema.Repository, storage models.GeneralRepository, rpc noderpc.INode, sharePath, network string, ipfs ...string) Parser {
	return Parser{
		bmdRepo: bmdRepo, blocksRepo: blocksRepo, storage: storage, protocolRepo: protocolRepo, schemaRepo: schemaRepo,
		rpc: rpc, sharePath: sharePath, network: network, ipfs: ipfs,
	}
}

// Parse -
func (t Parser) Parse(address string, level int64) ([]tokenmetadata.TokenMetadata, error) {
	state, err := t.getState(level)
	if err != nil {
		return nil, err
	}
	return t.parse(address, state)
}

// ParseBigMapDiff -
func (t Parser) ParseBigMapDiff(bmd *bigmapdiff.BigMapDiff) ([]tokenmetadata.TokenMetadata, error) {
	state, err := t.getState(bmd.Level)
	if err != nil {
		return nil, err
	}
	return t.parseBigMapDiff(bmd, state)
}

func (t Parser) parseBigMapDiff(bmd *bigmapdiff.BigMapDiff, state block.Block) ([]tokenmetadata.TokenMetadata, error) {
	if _, err := t.getBigMapPtr(bmd.Address, state); err != nil {
		return nil, err
	}

	m := &TokenMetadata{}
	value := gjson.Parse(bmd.Value)
	if err := m.Parse(value, bmd.Address, bmd.Ptr); err != nil {
		return nil, nil
	}
	m.Timestamp = bmd.Timestamp
	m.Level = bmd.Level

	if m.Link != "" {
		s := tzipStorage.NewFull(t.bmdRepo, t.blocksRepo, t.schemaRepo, t.storage, t.rpc, t.ipfs...)

		remoteMetadata := &TokenMetadata{}
		if err := s.Get(t.network, bmd.Address, m.Link, bmd.Ptr, remoteMetadata); err != nil {
			switch {
			case errors.Is(err, tzipStorage.ErrHTTPRequest):
				logger.Error(err)
				return nil, nil
			case errors.Is(err, tzipStorage.ErrNoIPFSResponse):
				remoteMetadata.Name = "Unknown"
			default:
				return nil, err
			}
		}
		m.Merge(remoteMetadata)
	}

	return []tokenmetadata.TokenMetadata{m.ToModel(bmd.Address, t.network)}, nil
}

func (t Parser) parse(address string, state block.Block) ([]tokenmetadata.TokenMetadata, error) {
	ptr, err := t.getBigMapPtr(address, state)
	if err != nil {
		return nil, err
	}

	bmd, err := t.bmdRepo.Get(bigmapdiff.GetContext{
		Ptr:          &ptr,
		Network:      t.network,
		Size:         1000, // TODO: max size
		CurrentLevel: &state.Level,
		Contract:     address,
	})
	if err != nil {
		return nil, err
	}

	if len(bmd) == 0 {
		return nil, nil
	}

	metadata := make([]*TokenMetadata, 0)
	for i := range bmd {
		m := &TokenMetadata{}
		value := gjson.Parse(bmd[i].Value)
		if err := m.Parse(value, address, ptr); err != nil {
			continue
		}
		m.Timestamp = bmd[i].Timestamp
		m.Level = bmd[i].Level
		metadata = append(metadata, m)
	}

	result := make([]tokenmetadata.TokenMetadata, 0)
	for _, m := range metadata {
		if m.Link != "" {
			s := tzipStorage.NewFull(t.bmdRepo, t.blocksRepo, t.schemaRepo, t.storage, t.rpc, t.ipfs...)

			remoteMetadata := &TokenMetadata{}
			if err := s.Get(t.network, address, m.Link, ptr, remoteMetadata); err != nil {
				if errors.Is(err, tzipStorage.ErrHTTPRequest) {
					logger.Error(err)
					return nil, nil
				}
				return nil, err
			}
			m.Merge(remoteMetadata)
		}

		result = append(result, m.ToModel(address, t.network))
	}

	return result, nil
}

func (t Parser) getState(level int64) (block.Block, error) {
	if level > 0 {
		return t.blocksRepo.Get(t.network, level)
	}
	return t.blocksRepo.Last(t.network)
}

func (t Parser) getBigMapPtr(address string, state block.Block) (int64, error) {
	storageMetadata, err := meta.GetSchema(t.schemaRepo, address, consts.STORAGE, state.Protocol)
	if err != nil {
		return 0, err
	}

	binPath := storageMetadata.Find(TokenMetadataStorageKey)
	if binPath == "" {
		return 0, ErrNoMetadataKeyInStorage
	}

	storageJSON, err := t.rpc.GetScriptStorageJSON(address, state.Level)
	if err != nil {
		return 0, err
	}

	ptrs, err := storage.FindBigMapPointers(storageMetadata, storageJSON)
	if err != nil {
		return 0, err
	}
	for ptr, path := range ptrs {
		if path == binPath {
			return ptr, nil
		}
	}

	return 0, ErrNoMetadataKeyInStorage
}
