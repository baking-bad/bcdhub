package storage

import (
	"encoding/json"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage/hash"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
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
	es      elastic.IElastic
	rpc     noderpc.INode
	network string
	address string
	ptr     int64
}

// NewTezosStorage -
func NewTezosStorage(es elastic.IElastic, rpc noderpc.INode, address, network string, ptr int64) TezosStorage {
	return TezosStorage{
		es:      es,
		rpc:     rpc,
		address: address,
		network: network,
		ptr:     ptr,
	}
}

// Get -
func (s TezosStorage) Get(value string) (*models.TZIP, error) {
	var uri TezosStorageURI
	if err := uri.Parse(value); err != nil {
		return nil, err
	}

	if err := uri.networkByChainID(s.es); err != nil {
		return nil, err
	}

	if err := s.fillFields(uri); err != nil {
		return nil, err
	}

	key, err := hash.Key(gjson.Parse(fmt.Sprintf(`{"string": "%s"}`, uri.Key)))
	if err != nil {
		return nil, err
	}

	bmd, err := s.es.GetBigMapKey(s.network, key, s.ptr)
	if err != nil {
		if elastic.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	decoded := DecodeValue(bmd.Value)

	var data models.TZIP
	err = json.Unmarshal([]byte(decoded), &data)
	return &data, err
}

func (s *TezosStorage) fillFields(uri TezosStorageURI) error {
	if uri.Network != "" {
		s.network = uri.Network
	}
	if uri.Address != "" && uri.Address != s.address {
		s.address = uri.Address

		block, err := s.es.GetLastBlock(s.network)
		if err != nil {
			return err
		}

		bmPtr, err := FindBigMapPointer(s.es, s.rpc, s.address, s.network, block.Protocol)
		if err != nil {
			return err
		}

		s.ptr = bmPtr
	}

	return nil
}

// FindBigMapPointer -
func FindBigMapPointer(es elastic.IElastic, rpc noderpc.INode, address, network, protocol string) (int64, error) {
	metadata, err := meta.GetMetadata(es, address, consts.STORAGE, protocol)
	if err != nil {
		return -1, err
	}
	binPath := metadata.Find(metadataAnnot)
	if binPath == "" {
		return -1, nil
	}
	registryStorage, err := rpc.GetScriptStorageJSON(address, 0)
	if err != nil {
		return -1, err
	}
	ptrs, err := storage.FindBigMapPointers(metadata, registryStorage)
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
