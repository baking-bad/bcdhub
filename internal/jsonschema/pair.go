package jsonschema

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/pkg/errors"
)

type pairMaker struct{}

func (m *pairMaker) Do(binPath string, metadata meta.Metadata) (Schema, error) {
	nm, ok := metadata[binPath]
	if !ok {
		return nil, errors.Errorf("[pairMaker] Unknown metadata binPath: %s", binPath)
	}
	schema := make(Schema)
	for _, arg := range nm.Args {
		subSchema, err := Create(arg, metadata)
		if err != nil {
			return nil, err
		}

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
	}, nil
}
