package main

import (
	"github.com/pkg/errors"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models/contract"
)

func getContract(ids []string) error {
	contracts := make([]contract.Contract, 0)
	if err := ctx.ES.GetByIDs(&contracts, ids...); err != nil {
		return errors.Errorf("[getContract] Find contracts error for IDs %v: %s", ids, err)
	}

	for i := range contracts {
		if err := parseContract(&contracts[i]); err != nil {
			return errors.Errorf("[getContract] Compute error message: %s", err)
		}
	}

	logger.Info("Metrics of %d contracts are computed", len(contracts))
	return ctx.ES.BulkUpdateField(contracts, "Alias", "Verified", "VerificationSource")
}

func parseContract(contract *contract.Contract) error {
	h := metrics.New(ctx.ES, ctx.DB)

	if _, err := h.SetContractAlias(contract); err != nil {
		return err
	}

	rpc, err := ctx.GetRPC(contract.Network)
	if err != nil {
		return err
	}
	if err = h.CreateTokenMetadata(rpc, ctx.SharePath, contract); err != nil {
		logger.Error(err)
	}

	return nil
}
