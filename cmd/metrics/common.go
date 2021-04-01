package main

import (
	"strconv"
	"strings"
)

func parseID(data []byte) (int64, error) {
	sData := string(data)
	id, err := strconv.ParseInt(sData, 10, 64)
	if err != nil {
		if strings.HasPrefix(sData, `"`) && strings.HasSuffix(sData, `"`) {
			sData = strings.TrimPrefix(sData, `"`)
			sData = strings.TrimSuffix(sData, `"`)
			return strconv.ParseInt(sData, 10, 64)
		}
		return id, err
	}
	return id, nil
}
