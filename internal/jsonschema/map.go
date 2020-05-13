package jsonschema

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
)

type mapMaker struct{}

func (m *mapMaker) Do(binPath string, metadata meta.Metadata) (Schema, DefaultModel, error) {
	nm, ok := metadata[binPath]
	if !ok {
		return nil, nil, fmt.Errorf("[mapMaker] Unknown metadata binPath: %s", binPath)
	}
	schema := Schema{
		"type":  "array",
		"title": nm.Prim,
	}
	if nm.Name != "" {
		schema["title"] = nm.Name
	}

	model := make(DefaultModel)
	keySchema, keyModel, err := Create(binPath+"/k", metadata)
	if err != nil {
		return nil, nil, err
	}
	model.Extend(keyModel, binPath+"/k")

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

	valueSchema, valueModel, err := Create(binPath+"/v", metadata)
	if err != nil {
		return nil, nil, err
	}
	model.Extend(valueModel, binPath+"/v")

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
	}, model, nil
}
