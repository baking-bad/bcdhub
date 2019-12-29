package contractparser

import (
	"strings"
)

var langPriorities = map[string]int{
	LangUnknown:   0,
	LangLigo:      1,
	LangLiquidity: 1,
	LangPython:    1,
}

func detectLiquidity(obj map[string]interface{}, entries []Entrypoint) bool {
	if annots, ok := obj["annots"]; ok {
		for _, a := range annots.([]interface{}) {
			s := a.(string)
			if strings.Contains(s, "_slash_") {
				return true
			}
		}
	}
	for i := range entries {
		if strings.Contains(entries[i].Name, "%_Liq_entry") {
			return true
		}
	}
	return false
}

func detectPython(obj map[string]interface{}) bool {
	if s, ok := obj["string"]; ok {
		str := s.(string)
		if strings.Contains(str, "https://SmartPy.io") {
			return true
		}
		if strings.Contains(str, "self.") {
			return true
		}
	}
	return false
}

func detectLIGO(obj map[string]interface{}) bool {
	if s, ok := obj["string"]; ok {
		if s.(string) == "GET_FORCE" {
			return true
		}
	}
	return false
}
