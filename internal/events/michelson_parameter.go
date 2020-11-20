package events

import (
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/tidwall/gjson"
)

// MichelsonParameterEvent -
type MichelsonParameterEvent struct {
	Sections
	Entrypoints []string

	name   string
	parser Parser
}

// NewMichelsonParameterEvent -
func NewMichelsonParameterEvent(event tzip.EventImplementation, name string) (*MichelsonParameterEvent, error) {
	parser, err := GetParser(name, event.MichelsonParameterEvent.ReturnType)
	if err != nil {
		return nil, err
	}
	return &MichelsonParameterEvent{
		Sections: Sections{
			Parameter:  event.MichelsonParameterEvent.Parameter,
			Code:       event.MichelsonParameterEvent.Code,
			ReturnType: event.MichelsonParameterEvent.ReturnType,
		},
		Entrypoints: event.MichelsonParameterEvent.Entrypoints,

		name:   name,
		parser: parser,
	}, nil
}

// Parse -
func (mpe *MichelsonParameterEvent) Parse(response gjson.Result) []TokenBalance {
	return mpe.parser.Parse(response)
}

// Normalize -
func (mpe *MichelsonParameterEvent) Normalize(parameters string) gjson.Result {
	p := gjson.Parse(parameters)
	if p.Get("value").Exists() {
		p = p.Get("value")
	}

	for prim := p.Get("prim").String(); prim == "Right" || prim == "Left"; prim = p.Get("prim").String() {
		p = p.Get("args.0")
	}
	return p
}
