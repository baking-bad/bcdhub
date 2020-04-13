package jsonschema

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
)

type orMaker struct{}

func (m *orMaker) Do(binPath string, metadata meta.Metadata) (Schema, error) {
	nm, ok := metadata[binPath]
	if !ok {
		return nil, fmt.Errorf("[orMaker] Unknown metadata binPath: %s", binPath)
	}

	schemas := make([]Schema, 0)
	for _, arg := range nm.Args {
		argSchema, err := Create(arg, metadata)
		if err != nil {
			return nil, err
		}

		subProperties := Schema{
			"schemaKey": Schema{
				"type":  "string",
				"const": arg,
			},
		}
		if properties, ok := argSchema["properties"]; ok {
			props := properties.(Schema)
			for k := range props {
				subProperties[k] = props[k]
			}
		}
		schemas = append(schemas, Schema{
			"title":      getOrTitile(arg, binPath),
			"properties": subProperties,
		})
	}

	name := nm.Name
	if nm.Name == "" {
		if nm.FieldName != "" {
			name = nm.FieldName
		} else {
			name = fmt.Sprintf("%s_%s", nm.Prim, strings.ReplaceAll(binPath, "/", ""))
		}
	}

	return Schema{
		"type":  "object",
		"title": name,
		"oneOf": schemas,
	}, nil
}

func getOrTitile(binPath, rootPath string) string {
	var result strings.Builder
	path := strings.TrimPrefix(binPath, rootPath)
	path = strings.TrimPrefix(path, "/")

	parts := strings.Split(path, "/")
	for i := range parts {
		if i != 0 {
			result.WriteByte(' ')
		}
		switch parts[i] {
		case "0":
			result.WriteString("Left")
		case "1":
			result.WriteString("Right")
		}
	}

	return result.String()
}
