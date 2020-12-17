package events

import (
	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/schema"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/tidwall/gjson"
)

// MichelsonExtendedStorage -
type MichelsonExtendedStorage struct {
	Sections

	name   string
	parser Parser

	protocol    string
	operationID string
	contract    string
	repo        schema.Repository
	bmd         []bigmapdiff.BigMapDiff
}

// NewMichelsonExtendedStorage -
func NewMichelsonExtendedStorage(impl tzip.EventImplementation, name, protocol, operationID, contract string, repo schema.Repository, bmd []bigmapdiff.BigMapDiff) (*MichelsonExtendedStorage, error) {
	parser, err := GetParser(name, impl.MichelsonExtendedStorageEvent.ReturnType)
	if err != nil {
		return nil, err
	}
	return &MichelsonExtendedStorage{
		Sections: Sections{
			Parameter:  impl.MichelsonExtendedStorageEvent.Parameter,
			Code:       impl.MichelsonExtendedStorageEvent.Code,
			ReturnType: impl.MichelsonExtendedStorageEvent.ReturnType,
		},

		name:        name,
		parser:      parser,
		protocol:    protocol,
		operationID: operationID,
		repo:        repo,
		bmd:         bmd,
		contract:    contract,
	}, nil
}

// Parse -
func (mes *MichelsonExtendedStorage) Parse(response gjson.Result) []TokenBalance {
	return mes.parser.Parse(response)
}

// Normalize - `value` is `Operation.DeffatedStorage`
func (mes *MichelsonExtendedStorage) Normalize(value string) gjson.Result {
	parser, err := contractparser.MakeStorageParser(nil, nil, mes.protocol, false)
	if err != nil {
		logger.Error(err)
		return gjson.Parse(value)
	}

	val, err := parser.Enrich(value, "", mes.bmd, true, false)
	if err != nil {
		logger.Error(err)
		return gjson.Parse(value)
	}

	metadata, err := meta.GetMetadata(mes.repo, mes.contract, consts.STORAGE, mes.protocol)
	if err != nil {
		logger.Error(err)
		return gjson.Parse(value)
	}

	val, err = storage.EnrichEmptyPointers(metadata, val)
	if err != nil {
		logger.Error(err)
		return gjson.Parse(value)
	}
	return val
}
