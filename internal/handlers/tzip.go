package handlers

import (
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/tzip"
	"github.com/pkg/errors"
)

// TZIP -
type TZIP struct {
	es      elastic.IElastic
	parsers map[string]tzip.Parser
}

// NewTZIP -
func NewTZIP(es elastic.IElastic, rpcs map[string]noderpc.INode, ipfs []string) *TZIP {
	parsers := make(map[string]tzip.Parser)
	for network, rpc := range rpcs {
		parsers[network] = tzip.NewParser(es, rpc, tzip.ParserConfig{
			IPFSGateways: ipfs,
		})
	}
	return &TZIP{
		es, parsers,
	}
}

// Do -
func (t *TZIP) Do(model elastic.Model) (bool, error) {
	bmd, ok := model.(*models.BigMapDiff)
	if !ok {
		return false, nil
	}
	if bmd.KeyHash != tzip.EmptyStringKey {
		return false, nil
	}
	return true, t.handle(bmd)
}

func (t *TZIP) handle(bmd *models.BigMapDiff) error {
	tzipParser, ok := t.parsers[bmd.Network]
	if !ok {
		return errors.Errorf("Unknown network for tzip parser: %s", bmd.Network)
	}

	model, err := tzipParser.Parse(tzip.ParseContext{
		BigMapDiff: *bmd,
	})
	if err != nil {
		logger.Error(err)
		return nil
	}
	if model == nil {
		return nil
	}

	logger.With(bmd).Info("Big map diff with TZIP is processed")
	return t.es.BulkInsert([]elastic.Model{model})
}
