package handlers

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	cmModel "github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	cm "github.com/baking-bad/bcdhub/internal/parsers/contract_metadata"
)

// ContractMetadata -
type ContractMetadata struct {
	repo    cmModel.Repository
	storage models.GeneralRepository
	parser  cm.Parser
}

// NewContractMetadata -
func NewContractMetadata(ctx *config.Context, ipfs []string) *ContractMetadata {
	return &ContractMetadata{
		ctx.ContractMetadata, ctx.Storage, cm.NewParser(ctx.BigMapDiffs, ctx.Blocks, ctx.Contracts, ctx.Storage, ctx.RPC, cm.ParserConfig{
			IPFSGateways: ipfs,
		}),
	}
}

// Do -
func (t *ContractMetadata) Do(ctx context.Context, bmd *domains.BigMapDiff, storage *ast.TypedAst) ([]models.Model, error) {
	if bmd.KeyHash != cm.EmptyStringKey {
		return nil, nil
	}
	return t.handle(ctx, bmd)
}

func (t *ContractMetadata) handle(ctx context.Context, bmd *domains.BigMapDiff) ([]models.Model, error) {
	model, err := t.parser.Parse(ctx, cm.ParseArgs{
		BigMapDiff: *bmd.BigMapDiff,
	})
	if err != nil {
		logger.Warning().Fields(bmd.LogFields()).Err(err).Msg("ContractMetadata.handle")
		return nil, nil
	}
	if model == nil {
		return nil, nil
	}

	m, err := t.repo.Get(model.Network, model.Address)
	if err == nil && m.OffChain {
		return nil, nil
	}

	return []models.Model{model}, nil
}
