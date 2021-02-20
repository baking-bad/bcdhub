package main

import (
	"github.com/pkg/errors"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
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
	return ctx.Contracts.UpdateField(contracts, "Alias", "Verified", "VerificationSource")
}

func parseContract(contract *contract.Contract) error {
	h := metrics.New(ctx.Contracts, ctx.BigMapDiffs, ctx.Blocks, ctx.Protocols, ctx.Operations, ctx.Schema, ctx.TokenBalances, ctx.TokenMetadata, ctx.TZIP, ctx.Migrations, ctx.Storage, ctx.DB)

	aliases, err := getAliases(contract.Network)
	if err != nil {
		return err
	}

	if _, err := h.SetContractAlias(contract, aliases); err != nil {
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

func getAliases(network string) (map[string]string, error) {
	if network != consts.Mainnet {
		return nil, nil
	}

	item, err := ctx.Cache.Fetch("aliases", ctx.AliasesCacheSeconds, func() (interface{}, error) {
		return ctx.TZIP.GetAliasesMap(network)
	})
	if err != nil {
		if !ctx.Storage.IsRecordNotFound(err) {
			return nil, err
		}
		return nil, nil
	}

	return item.Value().(map[string]string), nil
}
