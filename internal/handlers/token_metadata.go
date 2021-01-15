package handlers

import (
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/schema"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/tzip/tokens"
	"github.com/pkg/errors"
)

// TokenMetadata -
type TokenMetadata struct {
	bulk    models.BulkRepository
	parsers map[string]tokens.Parser
}

// NewTokenMetadata -
func NewTokenMetadata(bigMapRepo bigmapdiff.Repository, blockRepo block.Repository, protocolRepo protocol.Repository, schemaRepo schema.Repository, storage models.GeneralRepository, bulk models.BulkRepository, rpcs map[string]noderpc.INode, sharePath string, ipfs []string) *TokenMetadata {
	parsers := make(map[string]tokens.Parser)
	for network, rpc := range rpcs {
		parsers[network] = tokens.NewParser(bigMapRepo, blockRepo, protocolRepo, schemaRepo, storage, rpc, sharePath, network, ipfs...)
	}
	return &TokenMetadata{
		bulk, parsers,
	}
}

// Do -
func (t *TokenMetadata) Do(model models.Model) (bool, error) {
	bmd, ok := model.(*bigmapdiff.BigMapDiff)
	if !ok {
		return false, nil
	}
	return true, t.handle(bmd)
}

func (t *TokenMetadata) handle(bmd *bigmapdiff.BigMapDiff) error {
	tokenParser, ok := t.parsers[bmd.Network]
	if !ok {
		return errors.Errorf("Unknown network for tzip parser: %s", bmd.Network)
	}

	tokenMetadata, err := tokenParser.Parse(bmd.Address, bmd.Level)
	if err != nil {
		if !errors.Is(err, tokens.ErrNoMetadataKeyInStorage) {
			logger.Error(err)
		}
		return nil
	}
	if tokenMetadata == nil {
		return nil
	}

	models := make([]models.Model, 0, len(tokenMetadata))
	for i := range tokenMetadata {
		logger.With(&tokenMetadata[i]).Info("Update of token metadata is found")
		models = append(models, &tokenMetadata[i])
	}
	return t.bulk.Insert(models)
}
