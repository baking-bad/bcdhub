package meta

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/tidwall/gjson"
)

// Entrypoint -
type Entrypoint struct {
	Name       string      `json:"miguel_name"`
	Type       string      `json:"miguel_type"`
	Path       string      `json:"miguel_path"`
	Parameters interface{} `json:"miguel_parameters"`
}

// IsComplexEntryRoot -
func (metadata Metadata) IsComplexEntryRoot() bool {
	root := metadata["0"]
	return len(root.Args) > 0 &&
		root.Prim == consts.OR &&
		(root.Type == consts.TypeUnion ||
			root.Type == consts.TypeEnum ||
			root.Type == consts.TypeNamedEnum ||
			root.Type == consts.TypeNamedUnion)
}

// GetEntrypoints returns contract entrypoints
func (metadata Metadata) GetEntrypoints() ([]Entrypoint, error) {
	root := metadata["0"]

	ep := make([]Entrypoint, 0)
	if metadata.IsComplexEntryRoot() {
		for i, arg := range root.Args {
			nm := metadata[arg]

			params, err := parseEntrypointArg(metadata, nm, arg)
			if err != nil {
				return nil, err
			}
			ep = append(ep, Entrypoint{
				Name:       nm.GetEntrypointName(i),
				Parameters: params,
				Type:       nm.Prim,
				Path:       arg,
			})
		}
	} else {
		params, err := parseEntrypointArg(metadata, root, "0")
		if err != nil {
			return nil, err
		}
		ep = append(ep, Entrypoint{
			Name:       root.GetEntrypointName(-1),
			Parameters: params,
			Type:       root.Prim,
			Path:       "0",
		})
	}
	return ep, nil
}

// GetByPath - returns entrypoint name by parameters node
func (metadata Metadata) GetByPath(node gjson.Result) (string, error) {
	var entrypoint string
	if node.Get("entrypoint").Exists() {
		entrypoint = node.Get("entrypoint").String()
		node = node.Get("value")
	}

	startPath := "0"
	if entrypoint != "" {
		for key, nm := range metadata { // TODO: there can probably be collisions, better enumerate root args (if or)
			if nm.FieldName == entrypoint {
				startPath = key
				break
			}
		}
	}
	path := getPath(node, startPath)
	eMeta, ok := metadata[path]
	if !ok {
		return entrypoint, fmt.Errorf("Invalid parameter: %s", node.String())
	}
	if eMeta.Name != "" {
		return eMeta.Name, nil
	}

	if entrypoint == "" || (entrypoint == "default" && path != "0") {
		if path == "0" {
			return "default", nil
		}

		root := metadata["0"]
		for i := range root.Args {
			if root.Args[i] == path {
				return eMeta.GetEntrypointName(i), nil
			}
		}
	}

	return entrypoint, nil
}

func parseEntrypointArg(metadata Metadata, nm *NodeMetadata, path string) (interface{}, error) {
	switch nm.Type {
	case consts.TypeNamedTuple, consts.TypeNamedUnion, consts.TypeNamedEnum:
		return parseEntrypointNamed(metadata, nm, path)
	case consts.TypeTuple, consts.TypeUnion:
		return parseEntrypointTuple(metadata, nm, path)
	case consts.LIST, consts.SET:
		return parseEntrypointList(metadata, nm, path)
	case consts.OPTION:
		return parseEntrypointOption(metadata, nm, path)
	case consts.CONTRACT, consts.LAMBDA:
		params := gjson.Parse(nm.Parameter)
		data, err := formatter.MichelineToMichelson(params, true, formatter.DefLineSize)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"miguel_type":       nm.Type,
			"miguel_parameters": data,
		}, nil
	default:
		return map[string]interface{}{
			"miguel_type": nm.Type,
		}, nil
	}
}

func parseEntrypointTuple(metadata Metadata, nm *NodeMetadata, path string) (interface{}, error) {
	tupleMeta := metadata[path]
	if len(tupleMeta.Args) > 0 {
		res := make([]interface{}, len(tupleMeta.Args))
		for i, arg := range tupleMeta.Args {
			value, err := parseEntrypointArg(metadata, metadata[arg], arg)
			if err != nil {
				return nil, err
			}
			res[i] = value
		}
		return res, nil
	}
	return map[string]interface{}{
		"miguel_type": nm.Type,
	}, nil
}

func parseEntrypointNamed(metadata Metadata, nm *NodeMetadata, path string) (interface{}, error) {
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

func parseEntrypointList(metadata Metadata, nm *NodeMetadata, path string) (interface{}, error) {
	p := fmt.Sprintf("%s/l", path)
	if nm.Type == consts.SET {
		p = fmt.Sprintf("%s/s", path)
	}
	listMeta := metadata[p]
	if helpers.StringInArray(listMeta.Type, []string{consts.TypeNamedTuple, consts.TypeTuple, consts.TypeEnum, consts.OPTION, consts.TypeNamedEnum, consts.TypeUnion, consts.TypeNamedUnion}) {
		value, err := parseEntrypointArg(metadata, listMeta, p)
		if err != nil {
			return nil, err
		}
		return value, nil
	}

	if len(listMeta.Args) > 0 {
		params := make([]interface{}, len(listMeta.Args))
		hasList := false
		for i, arg := range listMeta.Args {
			value, err := parseEntrypointArg(metadata, metadata[arg], arg)
			if err != nil {
				return nil, err
			}
			params[i] = value

			if !hasList {
				hasList = metadata[arg].Type == consts.LIST
			}
		}
		if hasList {
			return params, nil
		}
		return map[string]interface{}{
			"miguel_type":       nm.Type,
			"miguel_parameters": params,
		}, nil
	}

	value, err := parseEntrypointArg(metadata, listMeta, p)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"miguel_type":       nm.Type,
		"miguel_parameters": []interface{}{value},
	}, nil
}

func parseEntrypointOption(metadata Metadata, nm *NodeMetadata, path string) (interface{}, error) {
	p := fmt.Sprintf("%s/o", path)
	optionMeta := metadata[p]
	if len(optionMeta.Args) > 0 {
		params := make([]interface{}, len(optionMeta.Args))
		for i, arg := range optionMeta.Args {
			value, err := parseEntrypointArg(metadata, metadata[arg], arg)
			if err != nil {
				return nil, err
			}
			params[i] = value
		}
		result := map[string]interface{}{
			"miguel_type":       nm.Type,
			"miguel_parameters": params,
		}
		return result, nil
	}
	value, err := parseEntrypointArg(metadata, optionMeta, p)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"miguel_type":       nm.Type,
		"miguel_parameters": []interface{}{value},
	}, nil
}

func getPath(node gjson.Result, path string) string {
	prim := node.Get("prim").String()
	if prim == "Left" {
		path += "/0"
		subNode := node.Get("args.0")
		return getPath(subNode, path)
	}

	if prim == "Right" {
		path += "/1"
		subNode := node.Get("args.0")
		return getPath(subNode, path)
	}

	if prim == "None" || prim == "Some" {
		return path + "/o"
	}
	return path
}
