package meta

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/translator"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// ParameterBuilder -
type ParameterBuilder struct {
	metadata          Metadata
	parameterBuilders map[string]parameterBuilderInterface
}

// NewParameterBuilder -
func NewParameterBuilder(metadata Metadata, needValidation bool) ParameterBuilder {
	b := ParameterBuilder{
		metadata: metadata,
	}
	b.parameterBuilders = map[string]parameterBuilderInterface{
		consts.PAIR: pairParameterBuilder{
			builder:  &b,
			metadata: metadata,
		},
		consts.OR: orParameterBuilder{
			builder:  &b,
			metadata: metadata,
		},
		consts.LIST: listParameterBuilder{
			builder:  &b,
			metadata: metadata,
		},
		consts.SET: listParameterBuilder{
			builder:  &b,
			metadata: metadata,
		},
		consts.OPTION: optionParameterBuilder{
			builder:  &b,
			metadata: metadata,
		},
		consts.MAP: mapParameterBuilder{
			builder:  &b,
			metadata: metadata,
		},
		consts.BIGMAP: mapParameterBuilder{
			builder:  &b,
			metadata: metadata,
		},
		consts.UNIT: unitParameterBuilder{},
		consts.LAMBDA: lambdaParameterBuilder{
			builder:  &b,
			metadata: metadata,
		},
		"default": defaultParameterBuilder{
			validate: needValidation,
		},
	}

	return b
}

// Build -
func (b ParameterBuilder) Build(binaryPath string, data map[string]interface{}) (gjson.Result, error) {
	binaryPath = b.prepareData(binaryPath, data)
	micheline, err := b.buildParameters(binaryPath, data)
	if err != nil {
		return gjson.Result{}, err
	}
	wrapped, err := b.wrapEntrypoint(binaryPath, micheline)
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.Parse(wrapped), nil
}

func (b ParameterBuilder) prepareData(binaryPath string, data map[string]interface{}) string {
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

func (b ParameterBuilder) buildParameters(path string, data map[string]interface{}) (string, error) {
	nm, ok := b.metadata[path]
	if !ok {
		return "", errors.Errorf("Unknown binary path: %s", path)
	}

	pb, ok := b.parameterBuilders[nm.Prim]
	if !ok {
		pb = b.parameterBuilders["default"]
	}
	return pb.Build(nm, path, data)
}

type parameterBuilderInterface interface {
	Build(node *NodeMetadata, path string, data map[string]interface{}) (string, error)
}

type defaultParameterBuilder struct {
	validate bool
}

func (b defaultParameterBuilder) Build(node *NodeMetadata, path string, data map[string]interface{}) (string, error) {
	value, ok := data[path]
	if !ok {
		if !helpers.StringInArray(node.Prim, []string{
			consts.STRING, consts.BYTES,
		}) {
			return "", errors.Errorf("'%s' is required", getName(node))
		}
		value = ""
	}

	if b.validate && !validate(node.Prim, value) {
		return "", errors.Errorf("Invalid %s \"%v\"", node.Prim, value)
	}

	switch node.Prim {
	case consts.STRING, consts.KEYHASH, consts.KEY, consts.ADDRESS, consts.CHAINID, consts.SIGNATURE, consts.CONTRACT:
		return fmt.Sprintf(`{"string": "%s"}`, value), nil
	case consts.BYTES:
		return fmt.Sprintf(`{"bytes": "%s"}`, value), nil
	case consts.INT, consts.NAT, consts.MUTEZ:
		switch t := value.(type) {
		case int, int64, int8, int32, int16, uint, uint16, uint32, uint64, uint8:
			return fmt.Sprintf(`{"int": "%d"}`, t), nil
		case float32, float64:
			return fmt.Sprintf(`{"int": "%0.f"}`, t), nil
		default:
			return "", errors.Errorf("[defaultBuilder] Invalid integer type: %v", t)
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
		return "", errors.Errorf("[defaultBuilder] Unknown primitive type: %s", node.Prim)
	}
}

type pairParameterBuilder struct {
	builder  *ParameterBuilder
	metadata Metadata
}

func (b pairParameterBuilder) Build(node *NodeMetadata, path string, data map[string]interface{}) (string, error) {
	s := ""
	for i, postfix := range []string{"/0", "/1"} {
		if i != 0 {
			s += ", "
		}
		argPath := path + postfix
		argStr, err := b.builder.buildParameters(argPath, data)
		if err != nil {
			return "", err
		}
		s += argStr
	}
	return fmt.Sprintf(`{"prim": "Pair", "args":[%s]}`, s), nil
}

type unitParameterBuilder struct{}

func (b unitParameterBuilder) Build(node *NodeMetadata, path string, data map[string]interface{}) (string, error) {
	return `{"prim": "Unit"}`, nil
}

type listParameterBuilder struct {
	builder  *ParameterBuilder
	metadata Metadata
}

func (b listParameterBuilder) Build(node *NodeMetadata, path string, data map[string]interface{}) (string, error) {
	value, ok := data[path]
	if !ok {
		return "", errors.Errorf("'%s' is required", getName(node))
	}
	listValue := interfaceSlice(value)

	listPath := path + "/l"
	if node.Prim == consts.SET {
		listPath = path + "/s"
	}

	var sb strings.Builder
	for i := range listValue {
		if i != 0 {
			sb.WriteByte(',')
		}

		switch t := listValue[i].(type) {
		case map[string]interface{}:
			argStr, err := b.builder.buildParameters(listPath, t)
			if err != nil {
				return "", err
			}
			sb.WriteString(argStr)
		default:
			data[listPath] = listValue[i]
			argStr, err := b.builder.buildParameters(listPath, data)
			if err != nil {
				return "", err
			}
			sb.WriteString(argStr)
		}
	}

	return fmt.Sprintf("[%s]", sb.String()), nil
}

type mapParameterBuilder struct {
	builder  *ParameterBuilder
	metadata Metadata
}

func (b mapParameterBuilder) Build(node *NodeMetadata, path string, data map[string]interface{}) (string, error) {
	value, ok := data[path]
	if !ok {
		return "", errors.Errorf("'%s' is required", getName(node))
	}
	var s string
	listValue := interfaceSlice(value)
	for i := range listValue {
		if i != 0 {
			s += ", "
		}
		mapValue, ok := listValue[i].(map[string]interface{})
		if !ok {
			return "", errors.Errorf("Invalid data: '%s'", getName(node))
		}
		var itemBuilder strings.Builder
		keyStr, err := b.builder.buildParameters(path+"/k", mapValue)
		if err != nil {
			return "", err
		}
		itemBuilder.WriteString(keyStr)
		itemBuilder.WriteByte(',')
		valStr, err := b.builder.buildParameters(path+"/v", mapValue)
		if err != nil {
			return "", err
		}
		itemBuilder.WriteString(valStr)
		s += fmt.Sprintf(`{"prim": "Elt", "args":[%s]}`, itemBuilder.String())
	}

	return fmt.Sprintf("[%s]", s), nil
}

type optionParameterBuilder struct {
	builder  *ParameterBuilder
	metadata Metadata
}

func (b optionParameterBuilder) Build(node *NodeMetadata, path string, data map[string]interface{}) (string, error) {
	value, ok := data[path]
	if !ok {
		return "", errors.Errorf("'%s' is required", getName(node))
	}
	mapValue, ok := value.(map[string]interface{})
	if !ok {
		return "", errors.Errorf("Invalid data: '%s'", getName(node))
	}
	schemaKey, ok := mapValue["schemaKey"]
	if !ok {
		return "", errors.Errorf("Invalid data: '%s'", getName(node))
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
		optionStr, err := b.builder.buildParameters(path+"/o", mapValue)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(`{"prim": "Some", "args":[%s]}`, optionStr), nil
	}
}

type orParameterBuilder struct {
	builder  *ParameterBuilder
	metadata Metadata
}

func (b orParameterBuilder) Build(node *NodeMetadata, path string, data map[string]interface{}) (string, error) {
	orData, ok := data[path]
	if !ok {
		return "", errors.Errorf("'%s' is required", getName(node))
	}
	mapValue, ok := orData.(map[string]interface{})
	if !ok {
		return "", errors.Errorf("Invalid data: '%s'", getName(node))
	}
	schemaKey, ok := mapValue["schemaKey"].(string)
	if !ok {
		return "", errors.Errorf("Invalid data: '%s'", getName(node))
	}
	if !strings.HasPrefix(schemaKey, path) {
		return "", errors.Errorf("Invalid data: '%s'", getName(node))
	}

	childStr, err := b.builder.buildParameters(schemaKey, mapValue)
	if err != nil {
		return "", err
	}

	orPath := strings.TrimPrefix(schemaKey, path+"/")
	return wrapLeftRight(orPath, childStr, false), nil
}

func (b ParameterBuilder) wrapEntrypoint(binPath, data string) (string, error) {
	nm, ok := b.metadata[binPath]
	if !ok {
		return "", errors.Errorf("Unknown binary path: %s", binPath)
	}
	return getParameterData(nm, binPath, data)
}

type lambdaParameterBuilder struct {
	builder  *ParameterBuilder
	metadata Metadata
}

func (b lambdaParameterBuilder) Build(node *NodeMetadata, path string, data map[string]interface{}) (string, error) {
	lambdaData, ok := data[path]
	if !ok {
		return "", errors.Errorf("'%s' is required", getName(node))
	}
	sLambda := fmt.Sprintf("%s", lambdaData)
	t, err := translator.NewConverter(
		translator.WithDefaultGrammar(),
	)
	if err != nil {
		return "", err
	}
	jsonLambda, err := t.FromString(sLambda)
	if err != nil {
		return "", err
	}
	return jsonLambda.String(), nil
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

func getParameterData(nm *NodeMetadata, binPath, data string) (string, error) {
	entrypoint := nm.Name
	if nm.Name == "" {
		entrypoint = "default"
		data = wrapLeftRight(binPath, data, true)
	}
	return fmt.Sprintf(`{"entrypoint": "%s", "value": %s}`, entrypoint, data), nil
}
