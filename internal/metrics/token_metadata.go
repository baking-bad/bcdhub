package metrics

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/tokens"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
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

	tokenMetadata, err := h.ES.GetTokenMetadatas(operation.Destination, operation.Network)
	if err != nil {
		if !elastic.IsRecordNotFound(err) {
			return err
		}
		return nil
	}
	registry := tokenMetadata[0].RegistryAddress

	parser := tokens.NewTokenMetadataParser(h.ES, rpc, sharePath, operation.Network)
	metadata, err := parser.ParseWithRegistry(operation.Destination, registry, operation.Level)
	if err != nil {
		return err
	}

	result := make([]elastic.Model, 0)
	for _, m := range metadata {
		newMetadata := m.ToModel(operation.Destination, operation.Network)
		for _, tm := range tokenMetadata {
			if newMetadata.Is(tm) {
				if newMetadata.Compare(tm) {
					newMetadata.ID = tm.ID
					result = append(result, newMetadata)
				}
				break
			}
		}
	}

	return h.ES.BulkUpdate(result)
}
