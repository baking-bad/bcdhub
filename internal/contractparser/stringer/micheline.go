package stringer

import (
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/valyala/fastjson"
)

// Micheline -
func Micheline(node gjson.Result) (gjson.Result, error) {
	if !node.IsArray() && !node.IsObject() {
		return node, nil
	}
	var p fastjson.Parser
	value, err := p.Parse(node.String())
	if err != nil {
		return node, err
	}

	newValue, err := processing(value)
	if err != nil {
		return node, err
	}

	return gjson.Parse(newValue.String()), nil
}

// MichelineFromBytes -
func MichelineFromBytes(data []byte) (gjson.Result, error) {
	var p fastjson.Parser
	value, err := p.ParseBytes(data)
	if err != nil {
		return gjson.Result{}, err
	}

	newValue, err := processing(value)
	if err != nil {
		return gjson.Result{}, err
	}

	return gjson.Parse(newValue.String()), nil
}

func processing(value *fastjson.Value) (*fastjson.Value, error) {
	switch value.Type() {
	case fastjson.TypeArray:
		return processingArray(value)
	case fastjson.TypeObject:
		return processingObject(value)
	case fastjson.TypeNull:
		return value, nil
	default:
		return value, errors.Errorf("Unknown node type: %s", value.Type().String())
	}
}

func processingArray(value *fastjson.Value) (*fastjson.Value, error) {
	array, err := value.Array()
	if err != nil {
		return nil, err
	}
	for i := range array {
		array[i], err = processing(array[i])
		if err != nil {
			return nil, err
		}
	}
	return value, nil
}

func processingObject(value *fastjson.Value) (*fastjson.Value, error) {
	if value.Exists("bytes") {
		valString := string(value.GetStringBytes("bytes"))
		unpackedValue := unpackBytes(valString)
		value.Del("bytes")
		arena := fastjson.Arena{}
		value.Set("string", arena.NewString(unpackedValue))
	} else if value.Exists("args") {
		args := value.Get("args")
		newArgs, err := processingArray(args)
		if err != nil {
			return nil, err
		}
		value.Set("args", newArgs)
	}
	return value, nil
}
