package events

import (
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
	"github.com/tidwall/gjson"
)

// MichelsonInitialStorage -
type MichelsonInitialStorage struct {
	Sections

	name   string
	parser tokenbalance.Parser
}

// NewMichelsonInitialStorage -
func NewMichelsonInitialStorage(impl tzip.EventImplementation, name string) (*MichelsonInitialStorage, error) {
	parser, err := tokenbalance.GetParser(name, impl.MichelsonInitialStorageEvent.ReturnType)
	if err != nil {
		return nil, err
	}
	return &MichelsonInitialStorage{
		Sections: Sections{
			Parameter:  impl.MichelsonInitialStorageEvent.Parameter,
			Code:       impl.MichelsonInitialStorageEvent.Code,
			ReturnType: impl.MichelsonInitialStorageEvent.ReturnType,
		},

		name:   name,
		parser: parser,
	}, nil
}

// Parse -
func (event *MichelsonInitialStorage) Parse(response gjson.Result) []tokenbalance.TokenBalance {
	balances := make([]tokenbalance.TokenBalance, 0)
	for _, item := range response.Get("storage").Array() {
		balance, err := event.parser.Parse(item)
		if err != nil {
			continue
		}
		balances = append(balances, balance)
	}
	return balances
}

// Normalize - `value` is `Operation.DeffatedStorage`
func (event *MichelsonInitialStorage) Normalize(value string) gjson.Result {
	return gjson.Parse(value)
}
