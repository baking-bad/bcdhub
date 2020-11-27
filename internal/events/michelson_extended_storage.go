package events

import (
	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
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
	es          elastic.IElastic
}

// NewMichelsonExtendedStorage -
func NewMichelsonExtendedStorage(impl tzip.EventImplementation, name, protocol, operationID string, es elastic.IElastic) (*MichelsonExtendedStorage, error) {
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
		es:          es,
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

	bmd, err := mes.es.GetBigMapDiffsUniqueByOperationID(mes.operationID)
	if err != nil {
		logger.Error(err)
		return gjson.Parse(value)
	}

	val, err := parser.Enrich(value, "", bmd, true)
	if err != nil {
		logger.Error(err)
		return gjson.Parse(value)
	}
	return val
}
