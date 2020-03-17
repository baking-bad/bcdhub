package miguel

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

var decoders = map[string]decoder{
	consts.TypeNamedTuple: &namedTupleDecoder{},
	consts.TypeTuple:      &tupleDecoder{},
	consts.LIST:           &listDecoder{},
	consts.SET:            &listDecoder{},
	consts.MAP:            &mapDecoder{},
	consts.BIGMAP:         &mapDecoder{},
	consts.TypeNamedUnion: &namedUnionDecoder{},
	consts.TypeUnion:      &namedUnionDecoder{},
	consts.OR:             &orDecoder{},
	consts.LAMBDA:         &lambdaDecoder{},
	consts.OPTION:         &optionDecoder{},
	"default":             newLiteralDecoder(),
}

// MichelineToMiguel -
func MichelineToMiguel(data gjson.Result, metadata meta.Metadata) (interface{}, error) {
	if !data.IsArray() && !data.IsObject() {
		return nil, nil
	}
	node, startPath, entrypoint, err := getStartPath(data, metadata)
	if err != nil {
		return nil, err
	}

	res, err := michelineNodeToMiguel(node, startPath, metadata, true)
	if err != nil {
		return nil, err
	}

	root := metadata["0"]
	if root.Prim == consts.OR {
		for _, arg := range root.Args {
			if arg == startPath {
				return map[string]interface{}{
					entrypoint: res,
				}, nil
			}
		}
	}
	return res, nil
}

// BigMapToMiguel -
func BigMapToMiguel(data gjson.Result, binPath string, metadata meta.Metadata) (interface{}, error) {
	return michelineNodeToMiguel(data, binPath, metadata, false)
}

func getStartPath(data gjson.Result, metadata meta.Metadata) (gjson.Result, string, string, error) {
	var entrypoint, value gjson.Result
	if data.IsArray() {
		entrypoint = data.Get("0.entrypoint")
		value = data.Get("0.value")
	} else if data.IsObject() {
		entrypoint = data.Get("entrypoint")
		value = data.Get("value")
	}

	if entrypoint.Exists() && value.Exists() {
		root := metadata["0"]
		if root.Prim != consts.OR && root.Type != consts.TypeNamedUnion && root.Type != consts.TypeNamedTuple {
			return value, "0", "", nil
		}
		for path, md := range metadata {
			if md.FieldName == entrypoint.String() {
				return value, path, entrypoint.String(), nil
			}
		}
		return value, "0", entrypoint.String(), nil
	}
	return data, "0", "", nil
}

func michelineNodeToMiguel(node gjson.Result, path string, metadata meta.Metadata, isRoot bool) (interface{}, error) {
	nm, ok := metadata[path]
	if !ok {
		return nil, fmt.Errorf("Unknown metadata path: %s", path)
	}

	if dec, ok := decoders[nm.Type]; ok {
		return dec.Decode(node, path, nm, metadata, isRoot)
	}
	return decoders["default"].Decode(node, path, nm, metadata, isRoot)
}

// GetGJSONPath -
func GetGJSONPath(path string) string {
	parts := strings.Split(path, "/")
	res := buildPathFromArray(parts)
	return strings.TrimSuffix(res, ".")
}

func buildPathFromArray(parts []string) (res string) {
	if len(parts) == 0 {
		return
	}

	for _, part := range parts {
		switch part {
		case "l", "s":
			res += "args.#."
		case "k":
			res += "#.args.0."
		case "v":
			res += "#.args.1."
		case "o":
			res += "args.0."
		default:
			res += fmt.Sprintf("args.%s.", part)
		}
	}
	return
}

func getGJSONPathUnion(path string, node gjson.Result) (res string, err error) {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		idx := len(parts)
		for i, part := range parts {
			switch part {
			case "0":
				if node.IsObject() {
					if node.Get(res+"prim").String() != "Left" {
						return "", fmt.Errorf("Invalid path")
					}
					res += "args.0."
				} else {
					res += "#(prim==\"Left\").args.0."
				}
			case "1":
				if node.IsObject() {
					if node.Get(res+"prim").String() != "Right" {
						return "", fmt.Errorf("Invalid path")
					}
					res += "args.0."
				} else {
					res += "#(prim==\"Right\").args.0."
				}
			case "o":
				if node.Get(res+"prim").String() != consts.None {
					res += "args.0."
				}
			default:
				idx = i + 1
				goto Break
			}
		}
	Break:
		res += buildPathFromArray(parts[idx:])
	}
	res = strings.TrimSuffix(res, ".")
	return
}
