package metrics

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/tzip/tokens"
)

// CreateTokenMetadata -
func (h *Handler) CreateTokenMetadata(rpc noderpc.INode, sharePath string, c *models.Contract) error {
	if !helpers.StringInArray(consts.FA2Tag, c.Tags) {
		return nil
	}

	parser := tokens.NewTokenMetadataParser(h.ES, rpc, sharePath, c.Network)
	metadata, err := parser.Parse(c.Address, c.Level)
	if err != nil {
		return err
	}

	result := make([]elastic.Model, 0)
	for i := range metadata {
		logger.Info("Token metadata for %s with token id %d found", c.Address, metadata[i].TokenID)
		result = append(result, metadata[i].ToModel(c.Address, c.Network))
	}

	return h.ES.BulkInsert(result)
}

// FixTokenMetadata -
func (h *Handler) FixTokenMetadata(rpc noderpc.INode, sharePath string, operation *models.Operation) error {
	if operation.Kind != consts.Transaction || operation.Status != consts.Applied || !strings.HasPrefix(operation.Destination, "KT") {
		return nil
	}

	contract, err := h.ES.GetContract(map[string]interface{}{
		"network": operation.Network,
		"address": operation.Destination,
	})
	if err != nil {
		if !elastic.IsRecordNotFound(err) {
			return err
		}
		return nil
	}

	if !helpers.StringInArray(consts.TokenMetadataRegistryTag, contract.Tags) {
		return nil
	}

	tokenMetadatas, err := h.ES.GetTokenMetadata(elastic.GetTokenMetadataContext{
		Contract: operation.Destination,
		Network:  operation.Network,
		TokenID:  -1,
	})
	if err != nil {
		if !elastic.IsRecordNotFound(err) {
			return err
		}
		return nil
	}
	result := make([]elastic.Model, 0)

	for _, tokenMetadata := range tokenMetadatas {
		parser := tokens.NewTokenMetadataParser(h.ES, rpc, sharePath, operation.Network)
		metadata, err := parser.ParseWithRegistry(tokenMetadata.RegistryAddress, operation.Level)
		if err != nil {
			return err
		}

		for _, m := range metadata {
			newMetadata := m.ToModel(tokenMetadata.Address, tokenMetadata.Network)
			if newMetadata.HasToken(tokenMetadata.Network, tokenMetadata.Address, tokenMetadata.TokenID) {
				result = append(result, newMetadata)
				break
			}
		}
	}
	if len(result) == 0 {
		return nil
	}

	return h.ES.BulkUpdate(result)
}
