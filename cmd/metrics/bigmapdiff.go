package main

import (
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/parsers/tzip"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

func getBigMapDiff(data amqp.Delivery) error {
	bmID := parseID(data.Body)

	bmd := models.BigMapDiff{ID: bmID}
	if err := ctx.ES.GetByID(&bmd); err != nil {
		return errors.Errorf("[getBigMapDiff] Find big map diff error: %s", err)
	}

	if err := parseBigMapDiff(bmd); err != nil {
		return errors.Errorf("[getBigMapDiff] Compute error message: %s", err)
	}
	return nil
}

func parseBigMapDiff(bmd models.BigMapDiff) error {
	switch bmd.KeyHash {
	case tzip.EmptyStringKey:
		return tzipHandler(bmd)
	}
	return nil
}

func tzipHandler(bmd models.BigMapDiff) error {
	rpc, err := ctx.GetRPC(bmd.Network)
	if err != nil {
		return err
	}
	tzipParser := tzip.NewParser(ctx.ES, rpc, tzip.ParserConfig{
		IPFSGateways: ctx.Config.IPFSGateways,
	})

	model, err := tzipParser.Parse(tzip.ParseContext{
		BigMapDiff: bmd,
	})
	if err != nil {
		return err
	}
	if model == nil {
		return nil
	}

	if err := ctx.ES.BulkInsert([]elastic.Model{model}); err != nil {
		return err
	}

	logger.Info("Big map diff with TZIP %s processed", bmd.ID)
	return nil
}
