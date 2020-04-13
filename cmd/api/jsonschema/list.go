package jsonschema

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
)

type listMaker struct{}

func (m *listMaker) Do(binPath string, metadata meta.Metadata) (Schema, error) {
	nm, ok := metadata[binPath]
	if !ok {
		return nil, fmt.Errorf("[listMaker] Unknown metadata binPath: %s", binPath)
	}
	schema := Schema{
		"type":  "array",
		"title": nm.Prim,
	}
	if nm.Name != "" {
		schema["title"] = nm.Name
	}

	path := binPath + "/l"
	if nm.Type == consts.SET {
		path = binPath + "/s"
	}

	required := make([]string, 0)
	propertiesItems := Schema{}

	listSchema, err := Create(path, metadata)
	if err != nil {
		return nil, err
	}

	if properties, ok := listSchema["properties"]; ok {
		props := properties.(Schema)
		for k := range props {
			propertiesItems[k] = props[k]
			required = append(required, k)
			schema["x-itemTitle"] = k
		}
	}

	schema["items"] = Schema{
		"type":       "object",
		"properties": propertiesItems,
		"required":   required,
	}

	name := fmt.Sprintf("%s_%s", nm.Prim, strings.ReplaceAll(binPath, "/", ""))
	return Schema{
		"type": "object",
		"properties": Schema{
			name: schema,
		},
	}, nil
}
