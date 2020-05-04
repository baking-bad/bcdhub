package jsonschema

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
)

func optionWrapper(schema Schema, binPath string, metadata meta.Metadata) (Schema, error) {
	if !strings.HasSuffix(binPath, "/o") {
		return nil, nil
	}
	nm, ok := metadata[binPath]
	if !ok {
		return nil, fmt.Errorf("[optionWrapper] Unknown metadata binPath: %s", binPath)
	}
	schemas := []Schema{
		{
			"title": consts.None,
			"properties": Schema{
				"schemaKey": Schema{
					"type":  "string",
					"const": consts.NONE,
				},
			},
		},
	}
	subProperties := Schema{
		"schemaKey": Schema{
			"type":  "string",
			"const": consts.SOME,
		},
	}
	if properties, ok := schema["properties"]; ok {
		props := properties.(Schema)
		for k := range props {
			subProperties[k] = props[k]
		}
	}
	schemas = append(schemas, Schema{
		"title":      consts.Some,
		"properties": subProperties,
	})

	name := nm.Name
	if nm.Name == "" {
		if nm.FieldName != "" {
			name = nm.FieldName
		} else {
			name = nm.Prim
		}
	}

	return Schema{
		"type":  "object",
		"title": fmt.Sprintf("%s (optional)", name),
		"oneOf": schemas,
		"x-props": Schema{
			"dense":    true,
			"outlined": true,
		},
	}, nil

}
