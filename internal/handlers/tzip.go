package handlers

import (
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/schema"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/tzip"
	"github.com/pkg/errors"
)

// TZIP -
type TZIP struct {
	bulk    models.BulkRepository
	parsers map[string]tzip.Parser
}

// NewTZIP -
func NewTZIP(bigMapRepo bigmapdiff.Repository, blockRepo block.Repository, schemaRepo schema.Repository, storage models.GeneralRepository, bulk models.BulkRepository, rpcs map[string]noderpc.INode, ipfs []string) *TZIP {
	parsers := make(map[string]tzip.Parser)
	for network, rpc := range rpcs {
		parsers[network] = tzip.NewParser(bigMapRepo, blockRepo, schemaRepo, storage, rpc, tzip.ParserConfig{
			IPFSGateways: ipfs,
		})
	}
	return &TZIP{
		bulk, parsers,
	}
}

// Do -
func (t *TZIP) Do(model models.Model) (bool, error) {
	bmd, ok := model.(*bigmapdiff.BigMapDiff)
	if !ok {
		return false, nil
	}
	if bmd.KeyHash != tzip.EmptyStringKey {
		return false, nil
	}
	return true, t.handle(bmd)
}

func (t *TZIP) handle(bmd *bigmapdiff.BigMapDiff) error {
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
	return t.bulk.Insert([]models.Model{model})
}
