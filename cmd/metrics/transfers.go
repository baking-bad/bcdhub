package main

import (
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

func getTransfer(data amqp.Delivery) error {
	transferID := parseID(data.Body)

	transfer := models.Transfer{ID: transferID}
	if err := ctx.ES.GetByID(&transfer); err != nil {
		return errors.Errorf("[getTransfer] Find transfer error: %s", err)
	}

	if err := parseTransfer(transfer); err != nil {
		return errors.Errorf("[getTransfer] Compute error message: %s", err)
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

	logger.Info("Transfer %s processed", transfer.ID)
	return nil
}
