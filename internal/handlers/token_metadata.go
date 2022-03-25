package handlers

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/parsers/contract_metadata/tokens"
	"github.com/pkg/errors"
)

// TokenMetadata -
type TokenMetadata struct {
	storage models.GeneralRepository
	parser  tokens.Parser
}

// NewTokenMetadata -
func NewTokenMetadata(ctx *config.Context, ipfs []string) *TokenMetadata {
	return &TokenMetadata{
		ctx.Storage, tokens.NewParser(ctx, ipfs...),
	}
}

// Do -
func (t *TokenMetadata) Do(ctx context.Context, bmd *domains.BigMapDiff, storage *ast.TypedAst) ([]models.Model, error) {
	tokenMetadata, err := t.parser.ParseBigMapDiff(ctx, bmd, storage)
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
