package jsonschema

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
)

type pairMaker struct{}

func (m *pairMaker) Do(binPath string, metadata meta.Metadata) (Schema, error) {
	nm, ok := metadata[binPath]
	if !ok {
		return nil, fmt.Errorf("[pairMaker] Unknown metadata binPath: %s", binPath)
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
		}
	}

	return Schema{
		"type":       "object",
		"properties": schema,
	}, nil
}
