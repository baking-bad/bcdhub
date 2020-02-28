package miguel

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

// Entrypoint -
type Entrypoint struct {
	Name       string      `json:"name"`
	Parameters interface{} `json:"parameters"`
}

// GetEntrypoints -
func GetEntrypoints(metadata meta.Metadata) ([]Entrypoint, error) {
	root := metadata["0"]

	ep := make([]Entrypoint, 0)
	if len(root.Args) > 0 && root.Prim == consts.OR && (root.Type == consts.TypeUnion || root.Type == consts.TypeNamedTuple || root.Type == consts.TypeNamedUnion) {
		for i, arg := range root.Args {
			nm := metadata[arg]

			params, err := parseEntrypointArg(metadata, nm, arg)
			if err != nil {
				return nil, err
			}
			ep = append(ep, Entrypoint{
				Name:       nm.GetName(i),
				Parameters: params,
			})
		}
	} else {
		params, err := parseEntrypointArg(metadata, root, "0")
		if err != nil {
			return nil, err
		}
		ep = append(ep, Entrypoint{
			Name:       root.GetName(-1),
			Parameters: params,
		})
	}
	return ep, nil
}

func parseEntrypointArg(metadata meta.Metadata, nm *meta.NodeMetadata, path string) (interface{}, error) {
	switch nm.Type {
	case consts.TypeNamedTuple, consts.TypeNamedUnion, consts.TypeNamedEnum:
		return parseEntrypointNamed(metadata, nm, path)
	case consts.LIST, consts.SET:
		return parseEntrypointList(metadata, nm, path)
	case consts.OPTION:
		return parseEntrypointOption(metadata, nm, path)
	case consts.CONTRACT, consts.LAMBDA:
		params := gjson.Parse(nm.Parameter).Value()
		return map[string]interface{}{
			"type":   nm.Type,
			"params": params,
		}, nil
	default:
		return map[string]interface{}{
			"type": nm.Type,
		}, nil
	}
}

func parseEntrypointNamed(metadata meta.Metadata, nm *meta.NodeMetadata, path string) (interface{}, error) {
	res := make(map[string]interface{})
	for _, arg := range nm.Args {
		argMeta := metadata[arg]

		value, err := parseEntrypointArg(metadata, argMeta, arg)
		if err != nil {
			return nil, err
		}
		res[argMeta.Name] = value
	}
	return res, nil
}

func parseEntrypointList(metadata meta.Metadata, nm *meta.NodeMetadata, path string) (interface{}, error) {
	p := fmt.Sprintf("%s/l", path)
	if nm.Type == consts.SET {
		p = fmt.Sprintf("%s/s", path)
	}
	listMeta := metadata[p]
	if len(listMeta.Args) > 0 {
		res := make([]interface{}, len(listMeta.Args))
		for i, arg := range listMeta.Args {
			value, err := parseEntrypointArg(metadata, metadata[arg], arg)
			if err != nil {
				return nil, err
			}
			res[i] = value
		}
		return res, nil
	}
	value, err := parseEntrypointArg(metadata, listMeta, p)
	if err != nil {
		return nil, err
	}
	return []interface{}{
		value,
	}, nil
}

func parseEntrypointOption(metadata meta.Metadata, nm *meta.NodeMetadata, path string) (interface{}, error) {
	p := fmt.Sprintf("%s/o", path)
	optionMeta := metadata[p]
	if len(optionMeta.Args) > 0 {
		res := make([]interface{}, len(optionMeta.Args))
		for i, arg := range optionMeta.Args {
			value, err := parseEntrypointArg(metadata, metadata[arg], arg)
			if err != nil {
				return nil, err
			}
			res[i] = value
		}
		return res, nil
	}
	value, err := parseEntrypointArg(metadata, optionMeta, p)
	if err != nil {
		return nil, err
	}
	return []interface{}{
		value,
	}, nil
}
