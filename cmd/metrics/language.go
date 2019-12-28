package main

import (
	"encoding/json"
	"errors"
	"github.com/jinzhu/gorm/dialects/postgres"
	"strings"
)

const (
	langPython    = "python"
	langLiquidity = "liquidity"
	langLigo      = "ligo"
	langUnknown   = "michelson"
)

func detectLanguage(script postgres.Jsonb) (string, error) {
	var s map[string]interface{}
	if err := json.Unmarshal(script.RawMessage, &s); err != nil {
		return langUnknown, err
	}

	code, ok := s["code"]
	if !ok {
		return langUnknown, errors.New("Can't `code` tag")
	}

	sections, ok := code.([]interface{})
	if !ok {
		return langUnknown, errors.New("Can't find sections")
	}

	if len(sections) != 3 {
		return langUnknown, errors.New("Invalid sections count")
	}

	if detectLiquidity(sections) {
		return langLiquidity, nil
	} else if detectPython(string(script.RawMessage)) {
		return langPython, nil
	} else if detectLIGO(string(script.RawMessage)) {
		return langLigo, nil
	}
	return langUnknown, nil
}

func findAnnot(entity interface{}, annot string) bool {
	switch t := entity.(type) {
	case []interface{}:
		for _, item := range t {
			if findAnnot(item, annot) {
				return true
			}
		}
	case map[string]interface{}:
		annots, ok := t["annots"]
		if ok {
			for _, a := range annots.([]interface{}) {
				s := a.(string)
				if strings.Contains(s, annot) {
					return true
				}
			}
		}
		args, ok := t["args"]
		if ok {
			return findAnnot(args, annot)
		}
	}
	return false
}

func detectLiquidity(sections []interface{}) bool {
	parameter := sections[0]
	code := sections[2]
	return findAnnot(parameter, "%_Liq_entry") || findAnnot(code, "_slash_")
}

func detectPython(script string) bool {
	if strings.Contains(script, "https://SmartPy.io") {
		return true
	}
	if strings.Contains(script, "self.") {
		return true
	}
	return false
}

func detectLIGO(script string) bool {
	return strings.Contains(script, "GET_FORCE")
}
