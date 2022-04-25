package storage

import (
	"context"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/types"
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
	ctx     *config.Context
	network types.Network
	address string
	ptr     int64
}

// TODO: multi-network
// NewTezosStorage -
func NewTezosStorage(ctx *config.Context, address string, ptr int64) TezosStorage {
	return TezosStorage{
		ctx:     ctx,
		address: address,
		ptr:     ptr,
	}
}

// Get -
func (s TezosStorage) Get(ctx context.Context, value string, output interface{}) error {
	var uri TezosStorageURI
	if err := uri.Parse(value); err != nil {
		return err
	}

	if err := uri.networkByChainID(); err != nil {
		if s.ctx.Storage.IsRecordNotFound(err) {
			return nil
		}
		return err
	}

	if err := s.fillFields(uri); err != nil {
		return err
	}

	key, err := ast.BigMapKeyHashFromString(fmt.Sprintf(`{"string": "%s"}`, uri.Key))
	if err != nil {
		return err
	}

	bmd, err := s.ctx.BigMapDiffs.Current(key, s.ptr)
	if err != nil {
		if s.ctx.Storage.IsRecordNotFound(err) {
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

		block, err := s.ctx.Blocks.Last()
		if err != nil {
			return err
		}

		bmPtr, err := storage.GetBigMapPtr(context.Background(), s.ctx.Storage, s.ctx.Contracts, s.ctx.RPC, s.address, metadataAnnot, block.Protocol.Hash, block.Level)
		if err != nil {
			return err
		}

		s.ptr = bmPtr
	}

	return nil
}
