package jsonschema

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/pkg/errors"
)

func optionWrapper(schema Schema, binPath string, metadata meta.Metadata) (Schema, error) {
	if !strings.HasSuffix(binPath, "/o") {
		return nil, nil
	}
	nm, ok := metadata[binPath]
	if !ok {
		return nil, errors.Errorf("[optionWrapper] Unknown metadata binPath: %s", binPath)
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
		}
	}

	return Schema{
		"type":    "object",
		"prim":    "option",
		"title":   name,
		"oneOf":   schemas,
		"default": consts.NONE,
	}, nil

}
