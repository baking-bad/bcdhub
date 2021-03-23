package main

import (
	"github.com/pkg/errors"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/parsers/tzip/tokens"
)

func getContract(ids []int64) error {
	contracts, err := ctx.Contracts.GetByIDs(ids...)
	if err != nil {
		return errors.Errorf("[getContract] Find contracts error for IDs %v: %s", ids, err)
	}

	for i := range contracts {
		if err := parseContract(&contracts[i]); err != nil {
			return errors.Errorf("[getContract] Compute error message: %s", err)
		}
	}

	logger.Info("Metrics of %d contracts are computed", len(contracts))

	items := make([]models.Model, len(contracts))
	for i := range contracts {
		items[i] = &contracts[i]
	}
	if err := saveSearchModels(ctx.Searcher, items); err != nil {
		return err
	}

	return ctx.Contracts.UpdateField(contracts, "Alias", "Verified", "VerificationSource", "ProjectID")
}

func parseContract(contract *contract.Contract) error {
	h := metrics.New(ctx.Contracts, ctx.BigMapDiffs, ctx.Blocks, ctx.Protocols, ctx.Operations, ctx.TokenBalances, ctx.TokenMetadata, ctx.TZIP, ctx.Migrations, ctx.Storage, ctx.DB)

	if contract.ProjectID == "" {
		if err := h.SetContractProjectID(contract); err != nil {
			return errors.Errorf("[parseContract] Error during set contract projectID: %s", err)
		}
	}

	rpc, err := ctx.GetRPC(contract.Network)
	if err != nil {
		return err
	}

	if err = h.CreateTokenMetadata(rpc, ctx.SharePath, contract, ctx.Config.IPFSGateways...); err != nil {
		if !errors.Is(err, tokens.ErrNoMetadataKeyInStorage) {
			logger.Error(err)
		}
	}

	return nil
}
