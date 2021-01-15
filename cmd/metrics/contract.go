package main

import (
	"github.com/pkg/errors"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/parsers/tzip/tokens"
)

func getContract(ids []string) error {
	contracts := make([]contract.Contract, 0)
	if err := ctx.Storage.GetByIDs(&contracts, ids...); err != nil {
		return errors.Errorf("[getContract] Find contracts error for IDs %v: %s", ids, err)
	}

	for i := range contracts {
		if err := parseContract(&contracts[i]); err != nil {
			return errors.Errorf("[getContract] Compute error message: %s", err)
		}
	}

	logger.Info("Metrics of %d contracts are computed", len(contracts))
	return ctx.Bulk.UpdateField(contracts, "Alias", "Verified", "VerificationSource")
}

func parseContract(contract *contract.Contract) error {
	h := metrics.New(ctx.Contracts, ctx.BigMapDiffs, ctx.Blocks, ctx.Protocols, ctx.Operations, ctx.Schema, ctx.TokenBalances, ctx.TokenMetadata, ctx.TZIP, ctx.Migrations, ctx.Storage, ctx.Bulk, ctx.DB)

	if _, err := h.SetContractAlias(contract); err != nil {
		return err
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
