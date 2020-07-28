package jsonschema

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
)

type orMaker struct{}

func (m *orMaker) Do(binPath string, metadata meta.Metadata) (Schema, DefaultModel, error) {
	nm, ok := metadata[binPath]
	if !ok {
		return nil, nil, fmt.Errorf("[orMaker] Unknown metadata binPath: %s", binPath)
	}
	switch nm.Type {
	case consts.TypeEnum, consts.TypeNamedEnum:
		return getEnum(binPath, metadata, nm)
	default:
		return getOr(binPath, metadata, nm)
	}
}

func getEnum(binPath string, metadata meta.Metadata, nm *meta.NodeMetadata) (Schema, DefaultModel, error) {
	oneOf := make([]Schema, 0)
	model := make(DefaultModel)
	for _, arg := range nm.Args {
		if _, ok := model[binPath]; !ok {
			model[binPath] = DefaultModel{
				"schemaKey": arg,
			}
		}
		oneOf = append(oneOf, Schema{
			"properties": Schema{
				"schemaKey": Schema{
					"type":  "string",
					"const": arg,
				},
			},
			"title": getOrTitle(arg, binPath, metadata),
		})
	}

	return Schema{
		"type":  "object",
		"prim":  nm.Prim,
		"title": getName(nm, binPath),
		"oneOf": oneOf,
	}, model, nil
}

func getOr(binPath string, metadata meta.Metadata, nm *meta.NodeMetadata) (Schema, DefaultModel, error) {
	oneOf := make([]Schema, 0)
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
		oneOf = append(oneOf, Schema{
			"title":      getOrTitle(arg, binPath, metadata),
			"properties": subProperties,
		})
	}

	return Schema{
		"type":  "object",
		"prim":  nm.Prim,
		"title": getName(nm, binPath),
		"oneOf": oneOf,
	}, model, nil
}

func getOrTitle(binPath, rootPath string, metadata meta.Metadata) string {
	var result strings.Builder
	nm, ok := metadata[binPath]
	if ok {
		if nm.Name != "" {
			return nm.Name
		}
		if nm.FieldName != "" {
			return nm.FieldName
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

	return result.String()
}

func getName(nm *meta.NodeMetadata, binPath string) string {
	if nm.Name != "" {
		return nm.Name
	}
	if nm.FieldName != "" {
		return nm.FieldName
	}
	return fmt.Sprintf("%s_%s", nm.Prim, strings.ReplaceAll(binPath, "/", ""))
}
