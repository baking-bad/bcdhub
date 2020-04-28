package jsonschema

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
)

type mapMaker struct{}

func (m *mapMaker) Do(binPath string, metadata meta.Metadata) (Schema, error) {
	nm, ok := metadata[binPath]
	if !ok {
		return nil, fmt.Errorf("[mapMaker] Unknown metadata binPath: %s", binPath)
	}
	schema := Schema{
		"type":  "array",
		"title": nm.Prim,
	}
	if nm.Name != "" {
		schema["title"] = nm.Name
	}

	keySchema, err := Create(binPath+"/k", metadata)
	if err != nil {
		return nil, err
	}

	required := make([]string, 0)
	propertiesItems := Schema{}
	if properties, ok := keySchema["properties"]; ok {
		props := properties.(Schema)
		for k := range props {
			propertiesItems[k] = props[k]
			required = append(required, k)
			schema["x-itemTitle"] = k
		}
	} else {
		propertiesItems[binPath+"/k"] = keySchema
	}

	valueSchema, err := Create(binPath+"/v", metadata)
	if err != nil {
		return nil, err
	}
	if properties, ok := valueSchema["properties"]; ok {
		props := properties.(Schema)
		for k := range props {
			propertiesItems[k] = props[k]
			required = append(required, k)
		}
	} else {
		propertiesItems[binPath+"/v"] = valueSchema
	}

	schema["items"] = Schema{
		"type":       "object",
		"properties": propertiesItems,
		"required":   required,
	}

	return Schema{
		"type": "object",
		"properties": Schema{
			binPath: schema,
		},
	}, nil
}
