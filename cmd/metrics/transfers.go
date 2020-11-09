package main

import (
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
)

func getTransfer(ids []string) error {
	transfers := make([]models.Transfer, 0)
	if err := ctx.ES.GetByIDs(&transfers, ids...); err != nil {
		return errors.Errorf("[getTransfer] Find transfer error for IDs %v: %s", ids, err)
	}

	for i := range transfers {
		if err := parseTransfer(transfers[i]); err != nil {
			return errors.Errorf("[getTransfer] Compute error message: %s", err)
		}
		logger.With(&transfers[i]).Info("Transfer is processed")
	}
	return nil
}

func parseTransfer(transfer models.Transfer) error {
	h := metrics.New(ctx.ES, ctx.DB)
	ok, err := h.SetTransferAliases(ctx.Aliases, &transfer)
	if err != nil {
		return err
	}
	if ok {
		if err := ctx.ES.UpdateFields(
			elastic.DocTransfers, transfer.ID,
			transfer,
			"FromAlias", "ToAlias", "Alias", "InitiatorAlias",
		); err != nil {
			return err
		}
	}

	return nil
}
