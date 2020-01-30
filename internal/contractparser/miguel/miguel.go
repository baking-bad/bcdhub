package miguel

import (
	"fmt"
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/meta"
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
	consts.OPTION:         newOptionDecoder(),
	"default":             newLiteralDecoder(),
}

// MichelineToMiguel -
func MichelineToMiguel(data gjson.Result, metadata meta.Metadata) (interface{}, error) {
	gjson.AddModifier("upper", func(json, arg string) string {
		return strings.ToUpper(json)
	})
	gjson.AddModifier("lower", func(json, arg string) string {
		return strings.ToLower(json)
	})

	node, startPath, err := getStartPath(data, metadata)
	if err != nil {
		return nil, err
	}

	return michelineNodeToMiguel(node, startPath, metadata)
}

func getStartPath(data gjson.Result, metadata meta.Metadata) (gjson.Result, string, error) {
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
		if root.Prim != consts.OR && root.Type != consts.TypeNamedUnion {
			return data, "", fmt.Errorf("Invalid root metadata: [prim] %s | [type] %s", root.Prim, root.Type)
		}
		for path, md := range metadata {
			if md.FieldName == entrypoint.String() {
				return value, path, nil
			}
		}
		return value, "0", nil
	}
	return data, "0", nil
}

func michelineNodeToMiguel(node gjson.Result, path string, metadata meta.Metadata) (interface{}, error) {
	nm, ok := metadata[path]
	if !ok {
		return nil, fmt.Errorf("Unknown metadata path: %s", path)
	}

	if dec, ok := decoders[nm.Type]; ok {
		return dec.Decode(node, path, nm, metadata)
	}
	return decoders["default"].Decode(node, path, nm, metadata)
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
		default:
			res += fmt.Sprintf("args.%s.", part)
		}
	}
	return
}

func getGJSONPathUnion(path string, node gjson.Result) (res string) {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		idx := len(parts)
		for i, part := range parts {
			switch part {
			case "0":
				if node.IsObject() {
					if node.Get(res+"prim").String() != "Left" {
						return ""
					}
					res += "args.0."
				} else {
					res += "#(prim==\"Left\").args.0."
				}
			case "1":
				if node.IsObject() {
					if node.Get(res+"prim").String() != "Right" {
						return ""
					}
					res += "args.0."
				} else {
					res += "#(prim==\"Right\").args.0."
				}
			case "o":
				res += "args.0."
			default:
				idx = i + 1
				break
			}
		}

		res += buildPathFromArray(parts[idx:])
	}
	res = strings.TrimSuffix(res, ".")
	return
}
