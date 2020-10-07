package migrations

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/parsers/tokens"
	"github.com/schollz/progressbar/v3"
)

type contract struct {
	Address string
	Network string
}

// CreateTokenMetadata -
type CreateTokenMetadata struct{}

// Key -
func (m *CreateTokenMetadata) Key() string {
	return "create_token_metadata"
}

// Description -
func (m *CreateTokenMetadata) Description() string {
	return "creates token metadata"
}

// Do - migrate function
func (m *CreateTokenMetadata) Do(ctx *config.Context) error {
	contracts, err := ctx.ES.GetContracts(map[string]interface{}{
		"tags": "fa2",
	})
	if err != nil {
		return err
	}
	logger.Info("Found %d contracts with tag 'fa2'", len(contracts))

	result := make([]elastic.Model, 0)
	registry := make(map[contract]struct{})

	bar := progressbar.NewOptions(len(contracts), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
	for i := range contracts {
		if err := bar.Add(1); err != nil {
			return err
		}
		if !helpers.StringInArray(tokens.TokenMetadataRegistry, contracts[i].Entrypoints) {
			continue
		}
		rpc, err := ctx.GetRPC(contracts[i].Network)
		if err != nil {
			return err
		}

		parser := tokens.NewTokenMetadataParser(ctx.ES, rpc, ctx.SharePath, contracts[i].Network)
		metadata, err := parser.Parse(contracts[i].Address, 0)
		if err != nil {
			continue
		}

		for j := range metadata {
			if metadata[j].IsEmpty() {
				continue
			}
			result = append(result, metadata[j].ToModel(contracts[i].Address, contracts[i].Network))
			registry[contract{
				Address: metadata[j].RegistryAddress,
				Network: contracts[i].Network,
			}] = struct{}{}
		}

	}

	dbTokens, err := ctx.DB.GetTokens()
	if err != nil {
		return err
	}

	for _, token := range dbTokens {
		result = append(result, &models.TokenMetadata{
			ID:        helpers.GenerateID(),
			Contract:  token.Contract,
			Network:   token.Network,
			Timestamp: time.Now(),
			TokenID:   int64(token.TokenID),
			Symbol:    token.Symbol,
			Name:      token.Name,
			Decimals:  int64(token.Decimals),
			Level:     1,
		})
	}

	if err := ctx.ES.BulkInsert(result); err != nil {
		logger.Errorf("ctx.ES.BulkInsert error: %v", err)
		return err
	}

	logger.Info("Done. %d token metadatas were saved.", len(result))

	logger.Info("Setting `token_metadata_registry` tag...")

	updates := make([]elastic.Model, 0)
	for c := range registry {
		contract, err := ctx.ES.GetContract(map[string]interface{}{
			"address": c.Address,
			"network": c.Network,
		})
		if err != nil {
			return err
		}
		contract.Tags = append(contract.Tags, consts.TokenMetadataRegistryTag)
		updates = append(updates, &contract)
	}

	if err := ctx.ES.BulkUpdate(updates); err != nil {
		logger.Errorf("ctx.ES.BulkUpdate error: %v", err)
		return err
	}
	logger.Info("Done. %d contracts were saved.", len(updates))

	return nil
}
