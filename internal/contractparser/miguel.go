package contractparser

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

// MichelineToMiguel -
func MichelineToMiguel(data gjson.Result, metadata Metadata) (interface{}, error) {
	node, startPath, err := getStartPath(data, metadata)
	if err != nil {
		return nil, err
	}
	return michelineToMiguelNode(node, startPath, metadata)
}

func getStartPath(data gjson.Result, metadata Metadata) (gjson.Result, string, error) {
	entrypoint := data.Get("0.entrypoint")
	value := data.Get("0.value")
	if data.IsObject() {
		entrypoint = data.Get("entrypoint")
		value = data.Get("value")
	}
	if entrypoint.Exists() && value.Exists() {
		root := metadata["0"]
		if root.Prim != OR && root.Type != TypeNamedUnion {
			return data, "", fmt.Errorf("Invalid root metadata: [prim] %s | [type] %s", root.Prim, root.Type)
		}
		for path, md := range metadata {
			if md.FieldName != entrypoint.String() {
				continue
			}
			return value, path, nil
		}
		return value, "0", nil
	}
	return data, "0", nil
}

func michelineToMiguelNode(node gjson.Result, path string, metadata Metadata) (interface{}, error) {
	nm, ok := metadata[path]
	if !ok {
		return nil, fmt.Errorf("Unknown metadata path: %s", path)
	}

	switch nm.Type {
	case TypeNamedTuple:
		return decodeNamedTuple(node, path, nm, metadata)
	case TypeTuple:
		return decodeTuple(node, path, nm, metadata)
	case LIST, SET:
		return decodeList(node, path, nm, metadata)
	case MAP, BIGMAP:
		return decodeMap(node, path, nm, metadata)
	case TypeNamedUnion:
		return decodeNamedUnion(node, path, nm, metadata)
	case OR:
		return decodeDefault(node, path, nm, metadata)
	default:
		return decodeLiteral(node, nm)
	}
}

// GetGJSONPath -
func GetGJSONPath(path string) string {
	res := ""
	parts := strings.Split(path, "/")
	for _, part := range parts {
		switch part {
		case "o":
			res += "#(prim==\"Some\").args.0."
		case "l", "s":
			res += "args.#."
		case "k":
			res += "#.args.0."
		default:
			res += fmt.Sprintf("args.%s.", part)
		}
	}
	if len(res) > 0 {
		res = res[:len(res)-1]
	}

	return res
}

func getGJSONPathUnion(path string, node gjson.Result) string {
	res := ""
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		idx := len(parts)
		for i, part := range parts[1:] {
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
			default:
				idx = i + 1
				break
			}
		}

		for _, part := range parts[idx:] {
			switch part {
			case "o":
				res += "args.0."
			case "l", "s":
				res += "args.#."
			case "k":
				res += "#.args.0."
			default:
				res += fmt.Sprintf("args.%s.", part)
			}
		}
		res = res[:len(res)-1]
	}
	return res
}

func decodeLiteral(node gjson.Result, metadata *NodeMetadata) (interface{}, error) {
	switch metadata.Type {
	case KEYHASH, BYTES, CONTRACT, NAT, MUTEZ, ADDRESS, STRING, TIMESTAMP, KEY, INT, SIGNATURE:
		data, err := decodeSimpleTypes(node)
		if err != nil {
			return nil, err
		}
		return data, nil
	case BOOL:
		return node.Get("prim").Bool(), nil
	}
	return nil, nil
}

func decodeSimpleTypes(node gjson.Result) (interface{}, error) {
	prim := node.Get("prim").String()
	if prim == None {
		return nil, nil
	}
	for k, v := range node.Map() {
		switch k {
		case STRING, BYTES:
			return v.String(), nil
		case INT:
			return v.Int(), nil
		default:
			return nil, fmt.Errorf("Unknown simple type: %s", k)
		}
	}
	return nil, nil
}

func decodeNamedTuple(node gjson.Result, path string, nm *NodeMetadata, metadata Metadata) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	for _, arg := range nm.Args {
		argPath := strings.TrimPrefix(arg, path+"/")
		gjsonPath := GetGJSONPath(argPath)
		argNode := node.Get(gjsonPath)

		if argNode.Exists() {
			data, err := michelineToMiguelNode(argNode, arg, metadata)
			if err != nil {
				return nil, err
			}
			res[metadata[arg].Name] = data
		} else {
			res[metadata[arg].Name] = nil
		}
	}
	return res, nil
}
func decodeMap(node gjson.Result, path string, nm *NodeMetadata, metadata Metadata) (map[string]interface{}, error) {
	if node.Get("int").Exists() {
		return map[string]interface{}{}, nil
	}
	res := make(map[string]interface{})
	gjsonPath := GetGJSONPath("k")
	keyNode := node.Get(gjsonPath)

	for i, k := range keyNode.Array() {
		key, err := michelineToMiguelNode(k, path+"/k", metadata)
		if err != nil {
			return nil, err
		}
		if key != nil {
			gjsonPath := fmt.Sprintf("%d.args.1", i)
			valNode := node.Get(gjsonPath)

			var value interface{}
			if valNode.Exists() {
				value, err = michelineToMiguelNode(valNode, path+"/v", metadata)
				if err != nil {
					return nil, err
				}
			}

			switch kv := key.(type) {
			case string:
				res[kv] = value
			case int, int64:
				s := fmt.Sprintf("%d", kv)
				res[s] = value
			}
		}
	}
	return res, nil
}

func decodeTuple(node gjson.Result, path string, nm *NodeMetadata, metadata Metadata) ([]interface{}, error) {
	res := make([]interface{}, 0)
	for _, arg := range nm.Args {
		argPath := strings.TrimPrefix(arg, path+"/")
		gjsonPath := GetGJSONPath(argPath)
		argNode := node.Get(gjsonPath)
		if argNode.Exists() {
			data, err := michelineToMiguelNode(argNode, arg, metadata)
			if err != nil {
				return nil, err
			}
			res = append(res, data)
		} else {
			res = append(res, nil)
		}
	}
	return res, nil
}

func decodeList(node gjson.Result, path string, nm *NodeMetadata, metadata Metadata) ([]interface{}, error) {
	res := make([]interface{}, 0)
	if len(node.Array()) > 0 {
		subPath := "/l"
		if nm.Type == SET {
			subPath = "/s"
		}
		for _, arg := range node.Array() {
			data, err := michelineToMiguelNode(arg, path+subPath, metadata)
			if err != nil {
				return nil, err
			}
			res = append(res, data)
		}
	}
	return res, nil
}

func decodeNamedUnion(node gjson.Result, path string, nm *NodeMetadata, metadata Metadata) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	for _, arg := range nm.Args {
		unionPath := getGJSONPathUnion(arg, node)
		argNode := node.Get(unionPath)
		if argNode.Exists() {
			data, err := michelineToMiguelNode(argNode, arg, metadata)
			if err != nil {
				return nil, err
			}
			argMetadata := metadata[arg]
			res[argMetadata.Name] = data
			return res, nil
		}
	}

	return nil, nil
}

func decodeDefault(node gjson.Result, path string, nm *NodeMetadata, metadata Metadata) (map[string]interface{}, error) {
	res := make(map[string]interface{})

	root := metadata["0"]
	for _, arg := range root.Args {
		if !strings.HasPrefix(arg, path) {
			continue
		}
		argPath := strings.TrimPrefix(arg, path)
		unionPath := getGJSONPathUnion(argPath, node)
		argNode := node.Get(unionPath)
		if argNode.Exists() {
			data, err := michelineToMiguelNode(argNode, arg, metadata)
			if err != nil {
				return nil, err
			}
			argMetadata := metadata[arg]
			res[argMetadata.Name] = data
			return res, nil
		}
	}

	return nil, nil
}
