package newmiguel

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/tidwall/gjson"
)

type miguel struct {
	decoders map[string]decoder
}

func newMiguel() *miguel {
	m := &miguel{}
	decoders := map[string]decoder{
		consts.TypeNamedEnum:  &enumDecoder{},
		consts.TypeEnum:       &enumDecoder{},
		consts.TypeNamedTuple: &namedTupleDecoder{parent: m},
		consts.TypeTuple:      &tupleDecoder{parent: m},
		consts.TypeNamedUnion: &namedUnionDecoder{parent: m},
		consts.TypeUnion:      &unionDecoder{parent: m},
		consts.LIST:           &listDecoder{parent: m},
		consts.SET:            &listDecoder{parent: m},
		consts.MAP:            &mapDecoder{parent: m},
		consts.BIGMAP:         &mapDecoder{parent: m},
		consts.OR:             &orDecoder{parent: m},
		consts.LAMBDA:         &lambdaDecoder{},
		consts.OPTION:         &optionDecoder{parent: m},
		"default":             newLiteralDecoder(),
	}
	m.decoders = decoders

	return m
}

// Convert -
func (m *miguel) Convert(data gjson.Result, path string, metadata meta.Metadata, isRoot bool) (node *Node, err error) {
	nm, ok := metadata[path]
	if !ok {
		return nil, fmt.Errorf("Unknown metadata path: %s", path)
	}

	if dec, ok := m.decoders[nm.Type]; ok {
		node, err = dec.Decode(data, path, nm, metadata, isRoot)
	} else {
		node, err = m.decoders["default"].Decode(data, path, nm, metadata, isRoot)
	}
	if err != nil {
		return
	}
	if strings.HasSuffix(path, "/o") {
		node.IsOption = true
	}
	return
}

// MichelineToMiguel -
func MichelineToMiguel(data gjson.Result, metadata meta.Metadata) (*Node, error) {
	return newMiguel().Convert(data, "0", metadata, true)
}

// BigMapToMiguel -
func BigMapToMiguel(data gjson.Result, binPath string, metadata meta.Metadata) (*Node, error) {
	return newMiguel().Convert(data, binPath, metadata, false)
}

// ParameterToMiguel -
func ParameterToMiguel(data gjson.Result, metadata meta.Metadata) (*Node, error) {
	if !data.IsArray() && !data.IsObject() {
		return nil, nil
	}
	node, startPath, err := getStartPath(data, metadata)
	if err != nil {
		return nil, err
	}
	node, startPath = getGJSONParameterPath(node, startPath)
	res, err := newMiguel().Convert(node, startPath, metadata, true)
	if err != nil {
		return nil, err
	}
	return res, nil
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
		if root.Prim != consts.OR && root.Type != consts.TypeNamedUnion && root.Type != consts.TypeNamedTuple {
			return value, "0", nil
		}
		for path, md := range metadata {
			if md.FieldName == entrypoint.String() {
				parentParts := strings.Split(path, "/")
				parentParts = parentParts[0 : len(parentParts)-1]
				parent := strings.Join(parentParts, "/")
				parentNode, ok := metadata[parent]
				if !ok {
					continue
				}
				if parentNode.Type != consts.OR {
					continue
				}
				return value, path, nil
			}
		}
		return value, "0", nil
	}
	return data, "0", nil
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

func getGJSONParameterPath(node gjson.Result, startPath string) (gjson.Result, string) {
	path := startPath
	prim := node.Get("prim").String()
	if prim == "Right" {
		path += "/1"
		right := node.Get("args.0")
		return getGJSONParameterPath(right, path)
	}
	if prim == "Left" {
		path += "/0"
		left := node.Get("args.0")
		return getGJSONParameterPath(left, path)
	}
	return node, path
}

// GetGJSONPathForData -
func GetGJSONPathForData(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}

	var res strings.Builder
	for _, part := range parts {
		switch part {
		case "l", "s":
			res.WriteString("#.")
		case "k":
			res.WriteString("#.args.0.")
		case "v":
			res.WriteString("#.args.1.")
		case "o":
			res.WriteString("args.0.")
		default:
			res.WriteString(fmt.Sprintf("args.%s.", part))
		}
	}
	return strings.TrimSuffix(res.String(), ".")
}
