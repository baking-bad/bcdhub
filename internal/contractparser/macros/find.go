package macros

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// FindMacros -
func FindMacros(script gjson.Result) (gjson.Result, error) {
	gjson.AddModifier("upper", func(json, arg string) string {
		return strings.ToUpper(json)
	})
	gjson.AddModifier("lower", func(json, arg string) string {
		return strings.ToLower(json)
	})

	codePath := ""
	for i, item := range script.Get("code").Array() {
		if item.Get("prim").String() == "code" {
			codePath = fmt.Sprintf("code.%d.args", i)
			break
		}
	}
	code := script.Get(codePath)
	result, err := walkForMacros(code, "", code.String())
	if err != nil {
		return script, err
	}

	resultJSON := gjson.Parse(result)
	res, err := sjson.Set(script.String(), codePath, resultJSON.Value())
	if err != nil {
		return script, err
	}
	return gjson.Parse(res), nil
}

func walkForMacros(script gjson.Result, jsonPath, textScript string) (result string, err error) {
	result = textScript
	if script.IsArray() {
		for i, item := range script.Array() {
			var itemResult string
			itemJSONPath := getIndexJSONPath(jsonPath, i)
			itemResult, err = walkForMacros(item, itemJSONPath, result)
			if err != nil {
				return
			}
			result = itemResult
		}
		return applyMacros(result, jsonPath, arrayMacros)
	} else if script.IsObject() {
		if !script.Get("prim").Exists() {
			items := make([]string, 0)
			for k := range script.Map() {
				items = append(items, k)
			}
		}

		args := script.Get("args")
		if args.Exists() {
			var argsResult string
			argsJSONPath := getArgsJSONPath(jsonPath)
			argsResult, err = walkForMacros(args, argsJSONPath, result)
			if err != nil {
				return
			}
			result = argsResult
		}
		return applyMacros(result, jsonPath, objectMacros)
	}
	return result, fmt.Errorf("Unknown script type: %v", script)
}

func applyMacros(json, jsonPath string, allMacros []macros) (res string, err error) {
	res = json
	for _, macros := range allMacros {
		data := gjson.Parse(res).Get(jsonPath)
		if macros.Find(data) {
			macros.Collapse(data)
			res, err = macros.Replace(res, jsonPath)
			if err != nil {
				return
			}
		}
	}
	return
}

func getIndexJSONPath(jsonPath string, index int) string {
	if jsonPath != "" {
		return fmt.Sprintf("%s.%d", jsonPath, index)
	}
	return fmt.Sprintf("%d", index)
}

func getArgsJSONPath(jsonPath string) string {
	if jsonPath != "" {
		return fmt.Sprintf("%s.args", jsonPath)
	}
	return "args"
}
