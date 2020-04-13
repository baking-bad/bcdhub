package meta

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/tidwall/gjson"
)

// BuildEntrypointMicheline -
func (m Metadata) BuildEntrypointMicheline(binaryPath string, data map[string]interface{}) (interface{}, error) {
	var builder strings.Builder
	nm, ok := m[binaryPath]
	if !ok {
		return "", fmt.Errorf("Unknown binary path: %s", binaryPath)
	}
	builder.WriteString(fmt.Sprintf(`{"entrypoint": "%s", "value": `, nm.GetEntrypointName(-1)))

	preprocessing(binaryPath, data)
	if err := build(m, binaryPath, data, &builder); err != nil {
		return "", err
	}
	builder.WriteString(`}`)
	value := gjson.Parse(builder.String()).Value()
	return value, nil
}

func build(metadata Metadata, path string, data map[string]interface{}, builder *strings.Builder) error {
	nm, ok := metadata[path]
	if !ok {
		return fmt.Errorf("Unknown binary path: %s", path)
	}
	switch nm.Prim {
	case consts.PAIR, consts.OR:
		return pairBuilder(metadata, path, data, builder)
	case consts.UNIT:
		return unitBuilder(metadata, path, data, builder)
	case consts.LIST, consts.SET:
		return listBuilder(metadata, path, data, builder)
	case consts.OPTION:
		return optionBuilder(metadata, path, data, builder)
	default:
		return defaultBuilder(metadata, path, data, builder)
	}
}

func defaultBuilder(metadata Metadata, path string, data map[string]interface{}, builder *strings.Builder) error {
	nm, ok := metadata[path]
	if !ok {
		return fmt.Errorf("Unknown binary path: %s", path)
	}
	value, ok := data[path]
	if !ok {
		return fmt.Errorf("'%s' is required field", getName(nm))
	}
	builder.WriteByte('{')
	switch nm.Prim {
	case consts.STRING, consts.KEYHASH, consts.KEY, consts.ADDRESS, consts.CHAINID, consts.SIGNATURE, consts.CONTRACT, consts.TIMESTAMP, consts.LAMBDA:
		builder.WriteString(fmt.Sprintf(`"string": "%s"`, value))
	case consts.BYTES:
		builder.WriteString(fmt.Sprintf(`"bytes": "%s"`, value))
	case consts.INT, consts.NAT, consts.MUTEZ, consts.BIGMAP:
		builder.WriteString(fmt.Sprintf(`"int": %0.f`, value))
	case consts.BOOL:
		sBool := "False"
		if tb, ok := value.(bool); ok && tb {
			sBool = "True"
		}
		builder.WriteString(fmt.Sprintf(`"prim": "%s"`, sBool))
	default:
		return fmt.Errorf("[defaultBuilder] Unknown primitive type")
	}
	builder.WriteByte('}')
	return nil
}

func pairBuilder(metadata Metadata, path string, data map[string]interface{}, builder *strings.Builder) error {
	builder.WriteString(`{"prim": "Pair", "args":[`)
	for i, postfix := range []string{"/0", "/1"} {
		argPath := path + postfix
		if err := build(metadata, argPath, data, builder); err != nil {
			return err
		}
		if i == 0 {
			builder.WriteByte(',')
		}
	}
	builder.WriteString(`]}`)
	return nil
}

func unitBuilder(metadata Metadata, path string, data map[string]interface{}, builder *strings.Builder) error {
	builder.WriteString(`{"prim": "Unit"}`)
	return nil
}

func listBuilder(metadata Metadata, path string, data map[string]interface{}, builder *strings.Builder) error {
	nm, ok := metadata[path]
	if !ok {
		return fmt.Errorf("Unknown binary path: %s", path)
	}
	value, ok := data[path]
	if !ok {
		return fmt.Errorf("'%s' is required field", getName(nm))
	}
	listValue := interfaceSlice(value)
	builder.WriteByte('[')

	listPath := path + "/l"
	if nm.Prim == consts.SET {
		listPath = path + "/s"
	}
	for i := range listValue {
		if i != 0 {
			builder.WriteByte(',')
		}
		data[listPath] = listValue[i]
		if err := build(metadata, listPath, data, builder); err != nil {
			return err
		}
	}

	builder.WriteByte(']')
	return nil
}

func preprocessing(binPath string, data map[string]interface{}) {
	if !strings.HasSuffix(binPath, "/o") {
		return
	}
	schemaKey, ok := data["schemaKey"]
	if !ok {
		return
	}
	optionData := map[string]interface{}{
		"schemaKey": schemaKey,
	}

	for k, v := range data {
		if !strings.HasPrefix(k, binPath) {
			continue
		}
		optionData[k] = v
	}
	data[binPath] = optionData
}

func optionBuilder(metadata Metadata, path string, data map[string]interface{}, builder *strings.Builder) error {
	optionPath := path + "/o"
	nm, ok := metadata[optionPath]
	if !ok {
		return fmt.Errorf("Unknown binary path: %s", optionPath)
	}
	value, ok := data[optionPath]
	if !ok {
		return fmt.Errorf("'%s' is required field", getName(nm))
	}
	mapValue, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Invalid data: '%s'", getName(nm))
	}
	schemaKey, ok := mapValue["schemaKey"]
	if !ok {
		return fmt.Errorf("Invalid data: '%s'", getName(nm))
	}
	switch schemaKey {
	case consts.NONE:
		builder.WriteString(`{"prim": "None"}`)
	default:
		builder.WriteString(`{"prim": "Some", "args":[`)
		for k, v := range mapValue {
			if k == "schemaKey" {
				continue
			}
			data[k] = v
		}
		if err := build(metadata, optionPath, data, builder); err != nil {
			return err
		}
		builder.WriteString(`]}`)
	}
	return nil
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
