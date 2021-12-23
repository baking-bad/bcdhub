package tokens

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/storage"
	tzipStorage "github.com/baking-bad/bcdhub/internal/parsers/tzip/storage"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// Parser -
type Parser struct {
	bmdRepo       bigmapdiff.Repository
	blocksRepo    block.Repository
	contractsRepo contract.Repository
	tmRepo        tokenmetadata.Repository
	storage       models.GeneralRepository

	rpc     noderpc.INode
	network types.Network
	ipfs    []string
}

// NewParser -
func NewParser(bmdRepo bigmapdiff.Repository, blocksRepo block.Repository, contractsRepo contract.Repository, tmRepo tokenmetadata.Repository, storage models.GeneralRepository, rpc noderpc.INode, network types.Network, ipfs ...string) Parser {
	return Parser{
		bmdRepo: bmdRepo, blocksRepo: blocksRepo, storage: storage,
		rpc: rpc, network: network, ipfs: ipfs, tmRepo: tmRepo,
		contractsRepo: contractsRepo,
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
func (t Parser) ParseBigMapDiff(bmd *domains.BigMapDiff, storageAST *ast.TypedAst) ([]tokenmetadata.TokenMetadata, error) {
	storageType := ast.TypedAst{
		Nodes: []ast.Node{ast.Copy(storageAST.Nodes[0])},
	}
	if err := storageType.SettleFromBytes(bmd.Operation.DeffatedStorage); err != nil {
		return nil, err
	}
	ptrs := storageType.FindBigMapByPtr()
	if bm, ok := ptrs[bmd.Ptr]; !ok || bm.GetName() != TokenMetadataStorageKey {
		return nil, nil
	}

	m := new(TokenMetadata)
	if err := m.Parse(gjson.ParseBytes(bmd.Value), bmd.Contract, bmd.Ptr); err != nil {
		return nil, nil
	}
	m.Timestamp = bmd.Timestamp
	m.Level = bmd.Level

	if m.Link != "" {
		ptr := bmd.Ptr
		switch {
		case strings.HasPrefix(m.Link, "ipfs://"):
			if found, err := t.tmRepo.GetOne(bmd.Network, bmd.Contract, m.TokenID); err == nil && found != nil {
				if link, ok := found.Extras[""]; ok && link == m.Link {
					return nil, nil
				}
			}
		case strings.HasPrefix(m.Link, "tezos-storage:"):
			bmPtr, err := storage.GetBigMapPtr(t.storage, t.contractsRepo, t.rpc, t.network, bmd.Contract, "metadata", bmd.Protocol.Hash, bmd.Level)
			if err != nil {
				return nil, err
			}
			ptr = bmPtr
		}

		s := tzipStorage.NewFull(t.bmdRepo, t.contractsRepo, t.blocksRepo, t.storage, t.rpc, t.ipfs...)

		remoteMetadata := new(TokenMetadata)
		if err := s.Get(t.network, bmd.Contract, m.Link, ptr, remoteMetadata); err != nil {
			switch {
			case errors.Is(err, tzipStorage.ErrHTTPRequest):
				logger.Warning().Str("url", m.Link).Str("kind", "token_metadata").Err(err).Msg("")
				return nil, nil
			case errors.Is(err, tzipStorage.ErrNoIPFSResponse):
				remoteMetadata.Name = consts.Unknown
				logger.Warning().Str("url", m.Link).Str("kind", "token_metadata").Err(err).Msg("")
			default:
				return nil, err
			}
		}
		m.Merge(remoteMetadata)
	}

	return []tokenmetadata.TokenMetadata{m.ToModel(bmd.Contract, t.network)}, nil
}

func (t Parser) parse(address string, state block.Block) ([]tokenmetadata.TokenMetadata, error) {
	ptr, err := storage.GetBigMapPtr(t.storage, t.contractsRepo, t.rpc, state.Network, address, TokenMetadataStorageKey, state.Protocol.Hash, state.Level)
	if err != nil {
		return nil, err
	}

	bmd, err := t.bmdRepo.Get(bigmapdiff.GetContext{
		Ptr:          &ptr,
		Network:      t.network,
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
		m := new(TokenMetadata)
		value := gjson.ParseBytes(bmd[i].Value)
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
			s := tzipStorage.NewFull(t.bmdRepo, t.contractsRepo, t.blocksRepo, t.storage, t.rpc, t.ipfs...)

			remoteMetadata := &TokenMetadata{}
			if err := s.Get(t.network, address, m.Link, ptr, remoteMetadata); err != nil {
				if errors.Is(err, tzipStorage.ErrHTTPRequest) {
					logger.Err(err)
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
