package models

import (
	"strings"

	"github.com/tidwall/gjson"
)

func getFoundBy(keys map[string]gjson.Result, categories []string) string {
	for _, category := range categories {
		name := strings.Split(category, "^")
		if len(name) == 0 {
			continue
		}
		if _, ok := keys[name[0]]; ok {
			return category
		}
	}

	for category := range keys {
		return category
	}

	return ""
}
