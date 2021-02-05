package storage

import (
	"encoding/json"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage/hash"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/schema"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
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
	schemaRepo schema.Repository
	storage    models.GeneralRepository

	rpc     noderpc.INode
	network string
	address string
	ptr     int64
}

// NewTezosStorage -
func NewTezosStorage(bigMapRepo bigmapdiff.Repository, blockRepo block.Repository, schemaRepo schema.Repository, storage models.GeneralRepository, rpc noderpc.INode, address, network string, ptr int64) TezosStorage {
	return TezosStorage{
		bigMapRepo: bigMapRepo,
		blockRepo:  blockRepo,
		schemaRepo: schemaRepo,
		storage:    storage,
		rpc:        rpc,
		address:    address,
		network:    network,
		ptr:        ptr,
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

	key, err := hash.Key(gjson.Parse(fmt.Sprintf(`{"string": "%s"}`, uri.Key)))
	if err != nil {
		return err
	}

	bmd, err := s.bigMapRepo.CurrentByKey(s.network, key, s.ptr)
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
		s.network = uri.Network
	}
	if uri.Address != "" && uri.Address != s.address {
		s.address = uri.Address

		block, err := s.blockRepo.Last(s.network)
		if err != nil {
			return err
		}

		bmPtr, err := FindBigMapPointer(s.schemaRepo, s.rpc, s.address, s.network, block.Protocol)
		if err != nil {
			return err
		}

		s.ptr = bmPtr
	}

	return nil
}

// FindBigMapPointer -
func FindBigMapPointer(schemaRepo schema.Repository, rpc noderpc.INode, address, network, protocol string) (int64, error) {
	metadata, err := meta.GetSchema(schemaRepo, address, consts.STORAGE, protocol)
	if err != nil {
		return -1, err
	}
	binPath := metadata.Find(metadataAnnot)
	if binPath == "" {
		return -1, nil
	}
	storageJSON, err := rpc.GetScriptStorageJSON(address, 0)
	if err != nil {
		return -1, err
	}
	ptrs, err := storage.FindBigMapPointers(metadata, storageJSON)
	if err != nil {
		return -1, err
	}
	bmPtr := int64(-1)
	for ptr, path := range ptrs {
		if path == binPath {
			bmPtr = ptr
		}
	}
	if bmPtr == -1 {
		err = errors.Wrap(ErrUnknownBigMapPointer, fmt.Sprintf("%s %s", network, address))
	}
	return bmPtr, err
}
