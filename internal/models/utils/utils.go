package utils

import (
	"strings"

	"github.com/tidwall/gjson"
)

// StringArray -
func StringArray(hit gjson.Result, tag string) []string {
	res := make([]string, 0)
	for _, t := range hit.Get(tag).Array() {
		res = append(res, t.String())
	}
	return res
}

// GetFoundBy -
func GetFoundBy(keys map[string]gjson.Result, categories []string) string {
	for _, category := range categories {
		name := strings.Split(category, "^")
		if len(name) == 0 {
			continue
		}
		if _, ok := keys[name[0]]; ok {
			return name[0]
		}
	}

	for category := range keys {
		return category
	}

	return ""
}
