package jsonschema

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
)

type orMaker struct{}

func (m *orMaker) Do(binPath string, metadata meta.Metadata) (Schema, DefaultModel, error) {
	nm, ok := metadata[binPath]
	if !ok {
		return nil, nil, fmt.Errorf("[orMaker] Unknown metadata binPath: %s", binPath)
	}

	schemas := make([]Schema, 0)
	model := make(DefaultModel)
	for _, arg := range nm.Args {
		argSchema, argModel, err := Create(arg, metadata)
		if err != nil {
			return nil, nil, err
		}

		model.Extend(argModel, arg)

		arg = strings.TrimSuffix(arg, "/o")

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
		} else {
			subProperties[arg] = argSchema
		}
		schemas = append(schemas, Schema{
			"title":      getOrTitile(arg, binPath, metadata),
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
		"x-props": Schema{
			"dense":    true,
			"outlined": true,
		},
	}, model, nil
}

func getOrTitile(binPath, rootPath string, metadata meta.Metadata) string {
	var result strings.Builder
	nm, ok := metadata[binPath]
	if ok {
		if nm.Name != "" {
			result.WriteString(fmt.Sprintf("%s (", nm.Name))
		} else if nm.FieldName != "" {
			result.WriteString(fmt.Sprintf("%s (", nm.FieldName))
		}
	}
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

	if ok {
		if nm.Name != "" || nm.FieldName != "" {
			result.WriteByte(')')
		}
	}
	return result.String()
}
