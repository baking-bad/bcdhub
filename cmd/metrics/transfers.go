package main

import (
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/pkg/errors"
)

func getTransfer(ids []string) error {
	transfers := make([]transfer.Transfer, 0)
	if err := ctx.ES.GetByIDs(&transfers, ids...); err != nil {
		return errors.Errorf("[getTransfer] Find transfer error for IDs %v: %s", ids, err)
	}

	for i := range transfers {
		if err := parseTransfer(transfers[i]); err != nil {
			return errors.Errorf("[getTransfer] Compute error message: %s", err)
		}
	}
	logger.Info("%d transfers are processed", len(transfers))
	return nil
}

func parseTransfer(transfer transfer.Transfer) error {
	h := metrics.New(ctx.ES, ctx.DB)

	if flag, err := h.SetTransferAliases(&transfer); flag {
		if err := ctx.ES.UpdateFields(
			elastic.DocTransfers, transfer.ID,
			transfer,
			"FromAlias", "ToAlias", "Alias", "InitiatorAlias",
		); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}
