package handlers

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	cmModel "github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	cm "github.com/baking-bad/bcdhub/internal/parsers/contract_metadata"
	"github.com/pkg/errors"
)

// ContractMetadata -
type ContractMetadata struct {
	repo    cmModel.Repository
	storage models.GeneralRepository
	parsers map[types.Network]cm.Parser
}

// NewContractMetadata -
func NewContractMetadata(bigMapRepo bigmapdiff.Repository, blockRepo block.Repository, contractsRepo contract.Repository, storage models.GeneralRepository, repo cmModel.Repository, rpcs map[types.Network]noderpc.INode, ipfs []string) *ContractMetadata {
	parsers := make(map[types.Network]cm.Parser)
	for network, rpc := range rpcs {
		parsers[network] = cm.NewParser(bigMapRepo, blockRepo, contractsRepo, storage, rpc, cm.ParserConfig{
			IPFSGateways: ipfs,
		})
	}
	return &ContractMetadata{
		repo, storage, parsers,
	}
}

// Do -
func (t *ContractMetadata) Do(bmd *domains.BigMapDiff, storage *ast.TypedAst) ([]models.Model, error) {
	if bmd.KeyHash != cm.EmptyStringKey {
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

	model, err := tzipParser.Parse(cm.ParseContext{
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
