package normalize

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Type - normalizes `typ` with combs
func Type(typ gjson.Result) (gjson.Result, error) {
	var m interface{}
	if err := json.Unmarshal([]byte(typ.Raw), &m); err != nil {
		return typ, err
	}
	if err := processType(m); err != nil {
		return typ, err
	}
	b, err := json.Marshal(m)
	if err != nil {
		return typ, err
	}
	return gjson.ParseBytes(b), nil
}

// Data - normalizes `data` with combs by `typ`. `typ` must be normalized.
func Data(data, typ gjson.Result) (gjson.Result, error) {
	var m interface{}
	if err := json.Unmarshal([]byte(data.Raw), &m); err != nil {
		return typ, err
	}
	newData, err := processValue(m, typ)
	if err != nil {
		return data, err
	}
	b, err := json.Marshal(newData)
	if err != nil {
		return typ, err
	}
	return gjson.ParseBytes(b), nil
}

// ScriptCode - gets code tag from `script` and normalize code type
func ScriptCode(script gjson.Result) (gjson.Result, error) {
	code, err := Type(script.Get("code"))
	if err != nil {
		return script, err
	}
	s, err := sjson.SetRaw(script.Raw, "code", code.Raw)
	if err != nil {
		return script, err
	}
	return gjson.Parse(s), nil
}

func processType(data interface{}) error {
	if data == nil {
		return errors.Wrap(ErrDataIsNil, "processType")
	}
	switch val := data.(type) {
	case map[string]interface{}:
		if p, ok := val["prim"]; ok {
			if prim, ok := p.(string); ok {
				switch prim {
				case consts.PAIR:
					return buildPairTree(val)
				case consts.CODE:
					return nil
				}
			}
		}
		if args, ok := val["args"]; ok {
			return processType(args)
		}
		return nil
	case []interface{}:
		for i := range val {
			if err := processType(val[i]); err != nil {
				return err
			}
		}
		return nil
	}
	return errors.Wrapf(ErrInvalidDataType, "[processType] %T", data)
}

func processValue(data interface{}, typ gjson.Result) (interface{}, error) {
	if data == nil {
		return nil, errors.Wrap(ErrDataIsNil, "processValue")
	}
	switch {
	case typ.IsObject():
		return processValueObject(data, typ)
	case typ.IsArray():
		return processValueArray(data, typ)
	}
	return nil, errors.Wrapf(ErrInvalidJSON, "[processValue] %s", typ.String())
}

func buildPairTreeValue(data interface{}) (interface{}, error) {
	if data == nil {
		return nil, errors.Wrap(ErrDataIsNil, "buildPairTreeValue")
	}
	switch val := data.(type) {
	case []interface{}:
		pair := map[string]interface{}{
			"prim": consts.Pair,
			"args": val,
		}
		return buildPairTreeValue(pair)
	case map[string]interface{}:
		prim, ok := val["prim"]
		if !ok || prim != consts.Pair {
			return nil, errors.Wrapf(ErrInvalidPrimitive, "%v", prim)
		}
		args, ok := val["args"]
		if !ok {
			return nil, errors.Wrapf(ErrArgsAreAbsent, consts.Pair)
		}
		argsArr := args.([]interface{})
		if len(argsArr) == 2 {
			return data, nil
		}
		resp, err := buildPair(argsArr, consts.Pair)
		if err != nil {
			return nil, err
		}
		merge(val, resp)
	}
	return data, nil
}

func buildPairTree(data map[string]interface{}) error {
	if data == nil {
		return errors.Wrap(ErrDataIsNil, "buildPairTree")
	}
	args, ok := data["args"]
	if !ok {
		return errors.Wrapf(ErrArgsAreAbsent, consts.PAIR)
	}
	argsArr := args.([]interface{})
	if len(argsArr) == 2 {
		return nil
	}
	val, err := buildPair(argsArr, consts.PAIR)
	if err != nil {
		return err
	}
	merge(data, val)
	return nil
}

func buildPair(data []interface{}, prim string) (map[string]interface{}, error) {
	if data == nil {
		return nil, errors.Wrap(ErrDataIsNil, "buildPair")
	}
	res := make(map[string]interface{})
	res["prim"] = prim
	if len(data) == 2 {
		res["args"] = data
		return res, nil
	}
	arg := data[0]
	argsMap, err := buildPair(data[1:], prim)
	if err != nil {
		return nil, err
	}
	res["args"] = []interface{}{
		arg, argsMap,
	}
	return res, nil
}

func merge(one, two map[string]interface{}) {
	if annots, ok := one["annots"]; ok {
		two["annots"] = annots
	}
	for k := range two {
		one[k] = two[k]
	}
}

func processValueObject(data interface{}, typ gjson.Result) (interface{}, error) {
	var newData interface{}
	prim := typ.Get("prim")
	if prim.Exists() {
		switch prim.String() {
		case consts.PAIR:
			res, err := buildPairTreeValue(data)
			if err != nil {
				return nil, err
			}
			newData = res
		case consts.LIST, consts.SET:
			return processListValue(data, typ)
		case consts.MAP:
			return processMapValue(data, typ)
		case consts.BIGMAP:
			return processBigMapValue(data, typ)
		case consts.OPTION:
			return processOptionValue(data, typ)
		case consts.OR:
			return processOrValue(data, typ)
		case consts.TICKET:
			return processTicketValue(data, typ)
		case consts.LAMBDA:
			return data, nil
		default:
			newData = data
		}
	} else {
		newData = data
	}

	m, ok := newData.(map[string]interface{})
	if !ok {
		return nil, errors.Wrapf(ErrInvalidDataType, "[processValueObject] %T", newData)
	}
	if args, ok := m["args"]; ok {
		newArgs, err := processValue(args, typ.Get("args"))
		if err != nil {
			return nil, err
		}
		m["args"] = newArgs
	}
	return newData, nil
}

func processValueArray(data interface{}, typ gjson.Result) (interface{}, error) {
	if data == nil {
		return nil, errors.Wrap(ErrDataIsNil, "processValueArray")
	}
	arr, ok := data.([]interface{})
	if !ok {
		return nil, errors.Wrapf(ErrInvalidDataType, "[processValueArray] %T", data)
	}
	typArr := typ.Array()
	if len(arr) != len(typArr) {
		return nil, errors.Wrapf(ErrInvalidArrayLength, "[processValueArray] data=%d != typ=%d", len(arr), len(typArr))
	}
	newArr := make([]interface{}, len(arr))
	for i, item := range typArr {
		newItem, err := processValue(arr[i], item)
		if err != nil {
			return nil, err
		}
		newArr[i] = newItem
	}
	return newArr, nil
}

func processListValue(data interface{}, typ gjson.Result) (interface{}, error) {
	if data == nil {
		return nil, errors.Wrap(ErrDataIsNil, "processListValue")
	}
	arr, ok := data.([]interface{})
	if !ok {
		return nil, errors.Wrapf(ErrInvalidDataType, "[processListValue] %T", data)
	}
	listItemType := typ.Get("args.0")
	for i := range arr {
		newItem, err := processValue(arr[i], listItemType)
		if err != nil {
			return nil, err
		}
		arr[i] = newItem
	}
	return arr, nil
}

func processOptionValue(data interface{}, typ gjson.Result) (interface{}, error) {
	if data == nil {
		return nil, errors.Wrap(ErrDataIsNil, "processOptionValue")
	}
	m := data.(map[string]interface{})
	if prim, ok := m["prim"]; !ok || prim != consts.Some {
		return data, nil
	}
	args, ok := m["args"]
	if !ok {
		return nil, errors.Wrap(ErrArgsAreAbsent, consts.OPTION)
	}
	a := args.([]interface{})
	optionType := typ.Get("args.0")
	newArgs := make([]interface{}, len(a))
	for i := range a {
		newArg, err := processValue(a[i], optionType)
		if err != nil {
			return nil, err
		}
		newArgs[i] = newArg
	}
	m["args"] = newArgs
	return m, nil
}

func processMapValue(data interface{}, typ gjson.Result) (interface{}, error) {
	if data == nil {
		return nil, errors.Wrap(ErrDataIsNil, "processMapValue")
	}

	switch val := data.(type) {
	case []interface{}:
		newArr := make([]interface{}, len(val))
		for i := range val {
			newItem, err := processValue(val[i], typ)
			if err != nil {
				return nil, err
			}
			newArr[i] = newItem
		}
		return newArr, nil
	case map[string]interface{}:
		args, ok := val["args"]
		if !ok {
			return nil, errors.Wrap(ErrArgsAreAbsent, consts.ELT)
		}
		a := args.([]interface{})
		newArr := make([]interface{}, len(val))
		for i := range a {
			t := typ.Get(fmt.Sprintf("args.%d", i))
			newItem, err := processValue(a[i], t)
			if err != nil {
				return nil, err
			}
			newArr[i] = newItem
		}
		val["args"] = newArr
		return val, nil
	default:
		return nil, errors.Wrapf(ErrInvalidDataType, "[processMapValue] %T", data)
	}
}

func processBigMapValue(data interface{}, typ gjson.Result) (interface{}, error) {
	if data == nil {
		return nil, errors.Wrap(ErrDataIsNil, "processBigMapValue")
	}
	switch val := data.(type) {
	case []interface{}:
		newData := make([]interface{}, len(val))
		for i := range val {
			item, err := processMapValue(val[i], typ)
			if err != nil {
				return nil, err
			}
			newData[i] = item
		}
		return newData, nil
	case map[string]interface{}:
		return data, nil
	}
	return nil, errors.Wrapf(ErrInvalidDataType, "[processBigMapValue] %T", data)
}

func processTicketValue(data interface{}, typ gjson.Result) (interface{}, error) {
	if data == nil {
		return nil, errors.Wrap(ErrDataIsNil, "processTicketValue")
	}
	switch val := data.(type) {
	case []interface{}:
		if len(val) != 3 {
			return nil, errors.Wrap(ErrInvalidArrayLength, consts.TICKET)
		}
		typArg := typ.Get("args.0")
		arg, err := processValue(val[1], typArg)
		if err != nil {
			return nil, err
		}
		val[1] = arg
		newData, err := buildPairTreeValue(val)
		if err != nil {
			return nil, err
		}
		return newData, nil
	case map[string]interface{}:
		p, ok := val["prim"]
		if !ok || p.(string) != consts.Pair {
			return nil, errors.Wrapf(ErrInvalidPrimitive, "[processTicketValue] %v", p)
		}
		argsData := val["args"]
		args := argsData.([]interface{})
		switch len(args) {
		case 2:
			return data, nil
		case 3:
			return processTicketValue(args, typ)
		default:
			return nil, errors.Wrap(ErrInvalidArrayLength, consts.TICKET)
		}
	}
	return nil, errors.Wrapf(ErrInvalidDataType, "[processBigMapValue] %T", data)
}

func processOrValue(data interface{}, typ gjson.Result) (interface{}, error) {
	if data == nil {
		return nil, errors.Wrap(ErrDataIsNil, "processOrValue")
	}

	m, ok := data.(map[string]interface{})
	if !ok {
		return nil, errors.Wrap(ErrInvalidDataType, consts.OR)
	}
	args := m["args"]
	a := args.([]interface{})
	if len(a) != 1 {
		return nil, errors.Wrap(ErrInvalidArrayLength, consts.LEFT)
	}
	var t gjson.Result

	prim := m["prim"]
	switch prim {
	case consts.Left:
		t = typ.Get("args.0")
	case consts.Right:
		t = typ.Get("args.1")
	default:
		return nil, errors.Wrapf(ErrInvalidPrimitive, "[processOrValue] %v", prim)
	}

	newArg, err := processValue(a[0], t)
	if err != nil {
		return nil, err
	}
	a[0] = newArg

	return m, nil
}
