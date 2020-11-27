package events

import (
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/tidwall/gjson"
)

// MichelsonInitialStorage -
type MichelsonInitialStorage struct {
	Sections

	name   string
	parser Parser
}

// NewMichelsonInitialStorage -
func NewMichelsonInitialStorage(impl tzip.EventImplementation, name string) (*MichelsonInitialStorage, error) {
	parser, err := GetParser(name, impl.MichelsonInitialStorageEvent.ReturnType)
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
func (event *MichelsonInitialStorage) Parse(response gjson.Result) []TokenBalance {
	return event.parser.Parse(response)
}

// Normalize - `value` is `Operation.DeffatedStorage`
func (event *MichelsonInitialStorage) Normalize(value string) gjson.Result {
	return gjson.Parse(value)
}
