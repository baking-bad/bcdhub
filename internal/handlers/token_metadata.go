package handlers

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/contract_metadata/tokens"
	"github.com/pkg/errors"
)

// TokenMetadata -
type TokenMetadata struct {
	storage models.GeneralRepository
	parsers map[types.Network]tokens.Parser
}

// NewTokenMetadata -
func NewTokenMetadata(bigMapRepo bigmapdiff.Repository, blockRepo block.Repository, contractsRepo contract.Repository, tm tokenmetadata.Repository, storage models.GeneralRepository, rpcs map[types.Network]noderpc.INode, ipfs []string) *TokenMetadata {
	parsers := make(map[types.Network]tokens.Parser)
	for network, rpc := range rpcs {
		parsers[network] = tokens.NewParser(bigMapRepo, blockRepo, contractsRepo, tm, storage, rpc, network, ipfs...)
	}
	return &TokenMetadata{
		storage, parsers,
	}
}

// Do -
func (t *TokenMetadata) Do(ctx context.Context, bmd *domains.BigMapDiff, storage *ast.TypedAst) ([]models.Model, error) {
	tokenParser, ok := t.parsers[bmd.Network]
	if !ok {
		return nil, errors.Errorf("Unknown network for tzip parser: %s", bmd.Network)
	}

	tokenMetadata, err := tokenParser.ParseBigMapDiff(ctx, bmd, storage)
	if err != nil {
		if !errors.Is(err, tokens.ErrNoMetadataKeyInStorage) {
			logger.Err(err)
		}
		return nil, nil
	}
	if len(tokenMetadata) == 0 {
		return nil, nil
	}

	models := make([]models.Model, 0, len(tokenMetadata))
	for i := range tokenMetadata {
		models = append(models, &tokenMetadata[i])
	}
	return models, nil
}
