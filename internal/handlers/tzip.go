package handlers

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/types"
	tzipModel "github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/tzip"
	"github.com/pkg/errors"
)

// ContractMetadata -
type ContractMetadata struct {
	repo    tzipModel.Repository
	parsers map[types.Network]tzip.Parser
}

// NewContractMetadata -
func NewContractMetadata(bigMapRepo bigmapdiff.Repository, blockRepo block.Repository, contractsRepo contract.Repository, storage models.GeneralRepository, repo tzipModel.Repository, rpcs map[types.Network]noderpc.INode, ipfs []string) *ContractMetadata {
	parsers := make(map[types.Network]tzip.Parser)
	for network, rpc := range rpcs {
		parsers[network] = tzip.NewParser(bigMapRepo, blockRepo, contractsRepo, storage, rpc, tzip.ParserConfig{
			IPFSGateways: ipfs,
		})
	}
	return &ContractMetadata{
		repo, parsers,
	}
}

// Do -
func (t *ContractMetadata) Do(bmd *domains.BigMapDiff, storage *ast.TypedAst) ([]models.Model, error) {
	if bmd.KeyHash != tzip.EmptyStringKey {
		return nil, nil
	}
	res, err := t.handle(bmd)
	return res, err
}

func (t *ContractMetadata) handle(bmd *domains.BigMapDiff) ([]models.Model, error) {
	tzipParser, ok := t.parsers[bmd.Network]
	if !ok {
		return nil, errors.Errorf("Unknown network for tzip parser: %s", bmd.Network)
	}

	model, err := tzipParser.Parse(tzip.ParseContext{
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
