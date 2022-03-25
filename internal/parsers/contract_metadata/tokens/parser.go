package tokens

import (
	"context"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	cmStorage "github.com/baking-bad/bcdhub/internal/parsers/contract_metadata/storage"
	"github.com/baking-bad/bcdhub/internal/parsers/storage"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// Parser -
type Parser struct {
	ctx  *config.Context
	ipfs []string
}

// NewParser -
func NewParser(ctx *config.Context, ipfs ...string) Parser {
	return Parser{
		ctx:  ctx,
		ipfs: ipfs,
	}
}

// Parse -
func (t Parser) Parse(ctx context.Context, address string, level int64) ([]tokenmetadata.TokenMetadata, error) {
	state, err := t.getState(level)
	if err != nil {
		return nil, err
	}
	return t.parse(ctx, address, state)
}

// ParseBigMapDiff -
func (t Parser) ParseBigMapDiff(ctx context.Context, bmd *domains.BigMapDiff, storageAST *ast.TypedAst) ([]tokenmetadata.TokenMetadata, error) {
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
			if found, err := t.ctx.TokenMetadata.GetOne(bmd.Contract, m.TokenID); err == nil && found != nil {
				if link, ok := found.Extras[""]; ok && link == m.Link {
					return nil, nil
				}
			}
		case strings.HasPrefix(m.Link, "tezos-storage:"):
			bmPtr, err := storage.GetBigMapPtr(ctx, t.ctx.Storage, t.ctx.Contracts, t.ctx.RPC, bmd.Contract, "metadata", bmd.Protocol.Hash, bmd.Level)
			if err != nil {
				return nil, err
			}
			ptr = bmPtr
		}

		s := cmStorage.NewFull(t.ctx, t.ipfs...)

		remoteMetadata := new(TokenMetadata)
		if err := s.Get(ctx, bmd.Contract, m.Link, ptr, remoteMetadata); err != nil {
			switch {
			case errors.Is(err, cmStorage.ErrHTTPRequest):
				logger.Warning().Str("url", m.Link).Str("kind", "token_metadata").Err(err).Msg("")
				return nil, nil
			case errors.Is(err, cmStorage.ErrNoIPFSResponse):
				remoteMetadata.Name = consts.Unknown
				logger.Warning().Str("url", m.Link).Str("kind", "token_metadata").Err(err).Msg("")
			default:
				return nil, err
			}
		}
		m.Merge(remoteMetadata)
	}

	return []tokenmetadata.TokenMetadata{m.ToModel(bmd.Contract)}, nil
}

func (t Parser) parse(ctx context.Context, address string, state block.Block) ([]tokenmetadata.TokenMetadata, error) {
	ptr, err := storage.GetBigMapPtr(ctx, t.ctx.Storage, t.ctx.Contracts, t.ctx.RPC, address, TokenMetadataStorageKey, state.Protocol.Hash, state.Level)
	if err != nil {
		return nil, err
	}

	bmd, err := t.ctx.BigMapDiffs.Get(bigmapdiff.GetContext{
		Ptr:          &ptr,
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
			s := cmStorage.NewFull(t.ctx, t.ipfs...)

			remoteMetadata := &TokenMetadata{}
			if err := s.Get(ctx, address, m.Link, ptr, remoteMetadata); err != nil {
				if errors.Is(err, cmStorage.ErrHTTPRequest) {
					logger.Err(err)
					return nil, nil
				}
				return nil, err
			}
			m.Merge(remoteMetadata)
		}

		result = append(result, m.ToModel(address))
	}

	return result, nil
}

func (t Parser) getState(level int64) (block.Block, error) {
	if level > 0 {
		return t.ctx.Blocks.Get(level)
	}
	return t.ctx.Blocks.Last()
}
