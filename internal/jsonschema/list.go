package jsonschema

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
)

type listMaker struct{}

func getItemsType(binPath string, metadata meta.Metadata) (string, error) {
	nm := metadata[binPath]

	switch nm.Prim {
	case consts.STRING, consts.BYTES, consts.KEYHASH, consts.KEY, consts.CHAINID, consts.SIGNATURE, consts.CONTRACT, consts.LAMBDA, consts.ADDRESS:
		return "string", nil
	default:
		return "object", nil
	}
}

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

	itemsType, err := getItemsType(path, metadata)
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
	} else {
		propertiesItems[path] = listSchema
	}

	schema["items"] = Schema{
		"type":       itemsType,
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
