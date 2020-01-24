package macros

import (
	"fmt"
	"log"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// FindMacros -
func FindMacros(script gjson.Result) (string, error) {
	r, result, err := walkForMacros(script, "", script.String())
	if err != nil {
		return "", err
	}

	log.Println(r)

	return result, nil
}

func walkForMacros(script gjson.Result, jsonPath, textScript string) (reg string, result string, err error) {
	result = textScript
	if script.IsArray() {
		reg += "("
		items := make([]string, 0)
		for i, item := range script.Array() {
			var itemReg, itemResult string
			itemJSONPath := getIndexJSONPath(jsonPath, i)
			itemReg, itemResult, err = walkForMacros(item, itemJSONPath, result)
			if err != nil {
				return
			}
			result = itemResult
			items = append(items, itemReg)
		}
		reg += strings.Join(items, ",")
		reg += ")"
	} else if script.IsObject() {
		prim := script.Get("prim")
		if prim.Exists() {
			reg += prim.String()
		} else {
			items := make([]string, 0)
			for k := range script.Map() {
				items = append(items, k)
			}
			reg += strings.Join(items, ",")
		}

		args := script.Get("args")
		if args.Exists() {
			var argsReg, argsResult string
			argsJSONPath := getArgsJSONPath(jsonPath)
			argsReg, argsResult, err = walkForMacros(args, argsJSONPath, result)
			if err != nil {
				return
			}
			result = argsResult
			reg += argsReg
		}
	} else {
		return reg, result, fmt.Errorf("Unknown script type: %v", script)
	}

	reg = strings.ToLower(reg)

	if jsonPath == "" {
		return
	}

	result, reg, err = replaceAllMacros(result, reg, jsonPath)
	return
}

func replaceAllMacros(result, reg, jsonPath string) (res, regular string, err error) {
	res = result
	regular = reg
	for _, macros := range allMacros {
		if !macros.Is(reg) {
			continue
		}
		data := gjson.Parse(result).Get(jsonPath)
		value := macros.Collapse(data)

		res, err = sjson.Set(res, jsonPath, value)
		if err != nil {
			return
		}
		prim := strings.ToLower(value["prim"].(string))
		log.Println(regular)
		regular = replacePrim(regular, macros.GetRegular(), prim)
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
