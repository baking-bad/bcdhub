package storage

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/storage"
)

const (
	metadataAnnot = "metadata"
)

// Tezos storage prefix
const (
	PrefixTezosStorage = "tezos-storage"
)

// TezosStorage -
type TezosStorage struct {
	bigMapRepo bigmapdiff.Repository
	blockRepo  block.Repository
	storage    models.GeneralRepository

	rpc       noderpc.INode
	network   types.Network
	address   string
	ptr       int64
	sharePath string
}

// NewTezosStorage -
func NewTezosStorage(bigMapRepo bigmapdiff.Repository, blockRepo block.Repository, storage models.GeneralRepository, rpc noderpc.INode, address, sharePath string, network types.Network, ptr int64) TezosStorage {
	return TezosStorage{
		bigMapRepo: bigMapRepo,
		blockRepo:  blockRepo,
		storage:    storage,
		rpc:        rpc,
		address:    address,
		network:    network,
		ptr:        ptr,
		sharePath:  sharePath,
	}
}

// Get -
func (s TezosStorage) Get(value string, output interface{}) error {
	var uri TezosStorageURI
	if err := uri.Parse(value); err != nil {
		return err
	}

	if err := uri.networkByChainID(s.blockRepo); err != nil {
		if !s.storage.IsRecordNotFound(err) {
			return err
		}
		return nil
	}

	if err := s.fillFields(uri); err != nil {
		return err
	}

	key, err := ast.BigMapKeyHashFromString(fmt.Sprintf(`{"string": "%s"}`, uri.Key))
	if err != nil {
		return err
	}

	bmd, err := s.bigMapRepo.Current(s.network, key, s.ptr)
	if err != nil {
		if s.storage.IsRecordNotFound(err) {
			return nil
		}
		return err
	}

	decoded, err := decodeData(bmd.Value)
	if err != nil {
		return err
	}

	return json.Unmarshal(decoded, output)
}

func (s *TezosStorage) fillFields(uri TezosStorageURI) error {
	if uri.Network != "" {
		s.network = types.NewNetwork(uri.Network)
	}
	if uri.Address != "" && uri.Address != s.address {
		s.address = uri.Address

		block, err := s.blockRepo.Last(s.network)
		if err != nil {
			return err
		}

		bmPtr, err := storage.GetBigMapPtr(s.rpc, s.network, s.address, metadataAnnot, block.Protocol.Hash, s.sharePath, block.Level)
		if err != nil {
			return err
		}

		s.ptr = bmPtr
	}

	return nil
}
