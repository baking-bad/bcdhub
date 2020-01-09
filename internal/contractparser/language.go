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

func detectLiquidity(node Node, entries []string) bool {
	for _, a := range node.Annotations {
		if strings.Contains(a, "_slash_") {
			return true
		}
	}

	for _, e := range entries {
		if strings.Contains(e, "_Liq_entry") {
			return true
		}
	}
	return false
}

func detectPython(node Node) bool {
	str := node.GetString()
	if str == "" {
		return false
	}

	return strings.Contains(str, "https://SmartPy.io") || strings.Contains(str, "self.")
}

func detectLIGO(node Node) bool {
	return node.GetString() == "GET_FORCE"
}
