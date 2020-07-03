package meta

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/tidwall/gjson"
)

// BuildEntrypointMicheline -
func (m Metadata) BuildEntrypointMicheline(binaryPath string, data map[string]interface{}, needValidate bool) (gjson.Result, error) {
	binaryPath = prepareData(binaryPath, data)

	micheline, err := m.buildParameters(binaryPath, data, needValidate)
	if err != nil {
		return gjson.Result{}, err
	}
	wrapped, err := wrapEntrypoint(binaryPath, micheline, m)
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.Parse(wrapped), nil
}

func prepareData(binaryPath string, data map[string]interface{}) string {
	if strings.HasSuffix(binaryPath, "/o") { // Hack for high-level option
		binaryPath = strings.TrimSuffix(binaryPath, "/o")
		option := make(map[string]interface{})
		for k, v := range data {
			option[k] = v
		}
		data[binaryPath] = option
	}

	for k, v := range data {
		if strings.HasSuffix(k, "/o") {
			newKey := strings.TrimSuffix(k, "/o")
			data[newKey] = v
		}
	}
	return binaryPath
}

func (m Metadata) buildParameters(path string, data map[string]interface{}, needValidate bool) (string, error) {
	nm, ok := m[path]
	if !ok {
		return "", fmt.Errorf("Unknown binary path: %s", path)
	}

	switch nm.Prim {
	case consts.PAIR:
		return m.pairParametersBuilder(nm, path, data, needValidate)
	case consts.OR:
		return m.orParametersBuilder(nm, path, data, needValidate)
	case consts.UNIT:
		return m.unitParametersBuilder(nm, path, data, needValidate)
	case consts.LIST, consts.SET:
		return m.listParametersBuilder(nm, path, data, needValidate)
	case consts.OPTION:
		return m.optionParametersBuilder(nm, path, data, needValidate)
	case consts.MAP:
		return m.mapParametersBuilder(nm, path, data, needValidate)
	default:
		return m.defaultParametersBuilder(nm, path, data, needValidate)
	}
}

func (m Metadata) defaultParametersBuilder(node *NodeMetadata, path string, data map[string]interface{}, needValidate bool) (string, error) {
	value, ok := data[path]
	if !ok {
		return "", fmt.Errorf("'%s' is required field", getName(node))
	}

	if needValidate && !validate(node.Prim, value) {
		return "", fmt.Errorf("Invalid parameter input: %s %v", node.Prim, value)
	}

	switch node.Prim {
	case consts.STRING, consts.KEYHASH, consts.KEY, consts.ADDRESS, consts.CHAINID, consts.SIGNATURE, consts.CONTRACT, consts.LAMBDA:
		return fmt.Sprintf(`{"string": "%s"}`, value), nil
	case consts.BYTES:
		return fmt.Sprintf(`{"bytes": "%s"}`, value), nil
	case consts.INT, consts.NAT, consts.MUTEZ, consts.BIGMAP:
		switch t := value.(type) {
		case int, int64, int8, int32, int16, uint, uint16, uint32, uint64, uint8:
			return fmt.Sprintf(`{"int": "%d"}`, t), nil
		case float32, float64:
			return fmt.Sprintf(`{"int": "%0.f"}`, t), nil
		default:
			return "", fmt.Errorf("[defaultBuilder] Invalid integer type: %v", t)
		}
	case consts.TIMESTAMP:
		ts, err := time.Parse(time.RFC3339, value.(string))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`{"int": "%d"}`, ts.UTC().Unix()), nil
	case consts.BOOL:
		sBool := "False"
		if tb, ok := value.(bool); ok && tb {
			sBool = "True"
		}
		return fmt.Sprintf(`{"prim": "%s"}`, sBool), nil
	default:
		return "", fmt.Errorf("[defaultBuilder] Unknown primitive type: %s", node.Prim)
	}
}

func (m Metadata) pairParametersBuilder(node *NodeMetadata, path string, data map[string]interface{}, needValidate bool) (string, error) {
	s := ""
	for i, postfix := range []string{"/0", "/1"} {
		if i != 0 {
			s += ", "
		}
		argPath := path + postfix
		argStr, err := m.buildParameters(argPath, data, needValidate)
		if err != nil {
			return "", err
		}
		s += argStr
	}
	return fmt.Sprintf(`{"prim": "Pair", "args":[%s]}`, s), nil
}

func (m Metadata) unitParametersBuilder(node *NodeMetadata, path string, data map[string]interface{}, needValidate bool) (string, error) {
	return `{"prim": "Unit"}`, nil
}

func (m Metadata) listParametersBuilder(node *NodeMetadata, path string, data map[string]interface{}, needValidate bool) (string, error) {
	value, ok := data[path]
	if !ok {
		return "", fmt.Errorf("'%s' is required field", getName(node))
	}
	listValue := interfaceSlice(value)

	listPath := path + "/l"
	if node.Prim == consts.SET {
		listPath = path + "/s"
	}

	var builder strings.Builder
	for i := range listValue {
		if i != 0 {
			builder.WriteByte(',')
		}

		switch t := listValue[i].(type) {
		case map[string]interface{}:
			argStr, err := m.buildParameters(listPath, t, needValidate)
			if err != nil {
				return "", err
			}
			builder.WriteString(argStr)
		default:
			data[listPath] = listValue[i]
			argStr, err := m.buildParameters(listPath, data, needValidate)
			if err != nil {
				return "", err
			}
			builder.WriteString(argStr)
		}
	}

	return fmt.Sprintf("[%s]", builder.String()), nil
}

func (m Metadata) mapParametersBuilder(node *NodeMetadata, path string, data map[string]interface{}, needValidate bool) (string, error) {
	value, ok := data[path]
	if !ok {
		return "", fmt.Errorf("'%s' is required field", getName(node))
	}
	var s string
	listValue := interfaceSlice(value)
	for i := range listValue {
		if i != 0 {
			s += ", "
		}
		mapValue, ok := listValue[i].(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("Invalid data: '%s'", getName(node))
		}
		var itemBuilder strings.Builder
		keyStr, err := m.buildParameters(path+"/k", mapValue, needValidate)
		if err != nil {
			return "", err
		}
		itemBuilder.WriteString(keyStr)
		itemBuilder.WriteByte(',')
		valStr, err := m.buildParameters(path+"/v", mapValue, needValidate)
		if err != nil {
			return "", err
		}
		itemBuilder.WriteString(valStr)
		s += fmt.Sprintf(`{"prim": "Elt", "args":[%s]}`, itemBuilder.String())
	}

	return fmt.Sprintf("[%s]", s), nil
}

func (m Metadata) optionParametersBuilder(node *NodeMetadata, path string, data map[string]interface{}, needValidate bool) (string, error) {
	value, ok := data[path]
	if !ok {
		return "", fmt.Errorf("'%s' is required field", getName(node))
	}
	mapValue, ok := value.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Invalid data: '%s'", getName(node))
	}
	schemaKey, ok := mapValue["schemaKey"]
	if !ok {
		return "", fmt.Errorf("Invalid data: '%s'", getName(node))
	}
	switch schemaKey {
	case consts.NONE:
		return `{"prim": "None"}`, nil
	default:
		for k, v := range mapValue {
			if k == "schemaKey" {
				continue
			}
			data[k] = v
		}
		optionStr, err := m.buildParameters(path+"/o", mapValue, needValidate)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`{"prim": "Some", "args":[%s]}`, optionStr), nil
	}
}

func (m Metadata) orParametersBuilder(node *NodeMetadata, path string, data map[string]interface{}, needValidate bool) (string, error) {
	orData, ok := data[path]
	if !ok {
		return "", fmt.Errorf("'%s' is required", getName(node))
	}
	mapValue, ok := orData.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Invalid data: '%s'", getName(node))
	}
	schemaKey, ok := mapValue["schemaKey"].(string)
	if !ok {
		return "", fmt.Errorf("Invalid data: '%s'", getName(node))
	}
	if !strings.HasPrefix(schemaKey, path) {
		return "", fmt.Errorf("Invalid data: '%s'", getName(node))
	}

	childStr, err := m.buildParameters(schemaKey, mapValue, needValidate)
	if err != nil {
		return "", err
	}

	orPath := strings.TrimPrefix(schemaKey, path+"/")
	return wrapLeftRight(orPath, childStr, false), nil
}

func wrapEntrypoint(binPath, data string, metadata Metadata) (string, error) {
	nm, ok := metadata[binPath]
	if !ok {
		return "", fmt.Errorf("Unknown binary path: %s", binPath)
	}
	entrypoint := getEntrypointName(nm)
	if entrypoint == "default" {
		data = wrapLeftRight(binPath, data, true)
	}
	return fmt.Sprintf(`{"entrypoint": "%s", "value": %s}`, entrypoint, data), nil
}

func wrapLeftRight(path, data string, skipFirst bool) string {
	parts := strings.Split(path, "/")

	if skipFirst {
		if len(parts) < 2 {
			return data
		}
		parts = parts[1:]
	}

	s := ""
	for i := range parts {
		var side string
		switch parts[i] {
		case "0":
			side = "Left"
		case "1":
			side = "Right"
		default:
			return s
		}

		if s == "" {
			s = `{"prim": "` + side + `", "args":[%s]}`
		} else {
			s = fmt.Sprintf(s, `{"prim": "`+side+`", "args":[%s]}`)
		}
	}
	return fmt.Sprintf(s, data)
}

func getName(nm *NodeMetadata) string {
	if nm.Name == "" {
		return nm.Prim
	}
	return nm.Name
}

func interfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("InterfaceSlice() given a non-slice type")
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}

func getEntrypointName(node *NodeMetadata) string {
	if node.Name != "" {
		return node.Name
	}
	return "default"
}
