package metrics

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/events"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/tzip/tokens"
	"github.com/pkg/errors"
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
		tzip := metadata[i].ToModel(c.Address, c.Network)
		logger.With(tzip).Info("Token metadata is found")
		result = append(result, tzip)

		transfers, err := h.executeInitialStorageEvent(rpc, tzip)
		if err != nil {
			return err
		}
		result = append(result, transfers...)
	}

	return h.ES.BulkInsert(result)
}

// FixTokenMetadata -
func (h *Handler) FixTokenMetadata(rpc noderpc.INode, sharePath string, contract *models.Contract, operation *models.Operation) error {
	if !operation.IsTransaction() || !operation.IsApplied() || !operation.IsCall() {
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

func (h *Handler) executeInitialStorageEvent(rpc noderpc.INode, tzip *models.TZIP) ([]elastic.Model, error) {
	ops, err := h.ES.GetOperations(map[string]interface{}{
		"destination": tzip.Address,
		"network":     tzip.Network,
		"kind":        consts.Origination,
	}, 1, false)
	if err != nil {
		if elastic.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(ops) != 1 {
		return nil, errors.Errorf("Invalid operations count: len(ops) [%d] != 1", len(ops))
	}

	protocol, err := h.ES.GetProtocol(tzip.Network, "", -1)
	if err != nil {
		return nil, err
	}

	origination := ops[0]
	for i := range tzip.Events {
		for j := range tzip.Events[i].Implementations {
			if !tzip.Events[i].Implementations[j].MichelsonInitialStorageEvent.Empty() {
				event, err := events.NewMichelsonInitialStorage(tzip.Events[i].Implementations[j], tzip.Events[i].Name)
				if err != nil {
					return nil, err
				}

				balances, err := events.Execute(rpc, event, events.Context{
					Network:                  tzip.Network,
					Parameters:               origination.DeffatedStorage,
					Source:                   origination.Source,
					Initiator:                origination.Initiator,
					Amount:                   origination.Amount,
					HardGasLimitPerOperation: protocol.Constants.HardGasLimitPerOperation,
				})
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return nil, nil
}
