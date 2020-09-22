package main

import "strings"

func parseID(data []byte) string {
	id := string(data)
	if strings.HasPrefix(id, `"`) && strings.HasSuffix(id, `"`) {
		id = strings.TrimPrefix(id, `"`)
		id = strings.TrimSuffix(id, `"`)
	}
	return id
}
