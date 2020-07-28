package jsonschema

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
)

type pairMaker struct{}

func (m *pairMaker) Do(binPath string, metadata meta.Metadata) (Schema, DefaultModel, error) {
	nm, ok := metadata[binPath]
	if !ok {
		return nil, nil, fmt.Errorf("[pairMaker] Unknown metadata binPath: %s", binPath)
	}
	schema := make(Schema)
	model := make(DefaultModel)
	for _, arg := range nm.Args {
		subSchema, subModel, err := Create(arg, metadata)
		if err != nil {
			return nil, nil, err
		}
		model.Extend(subModel, arg)

		if properties, ok := subSchema["properties"]; ok {
			props := properties.(Schema)
			for k := range props {
				schema[k] = props[k]
			}
		} else {
			schema[arg] = subSchema
		}
	}

	return Schema{
		"type":       "object",
		"properties": schema,
	}, model, nil
}
