package storage

import (
	"encoding/hex"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/stringer"
	"github.com/tidwall/gjson"
)

var (
	defaultTimeout = time.Second * 10
)

// DecodeValue -
func DecodeValue(value string) string {
	jsonValue := gjson.Parse(value)
	if !jsonValue.Get("bytes").Exists() {
		return ""
	}

	values := stringer.Get(value)
	if len(values) != 0 {
		return sanitizeString(values[0])
	}

	decoded, _ := hex.DecodeString(jsonValue.Get("bytes").String())
	return sanitizeString(string(decoded))
}

// sanitizeString -
func sanitizeString(s string) string {
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") && len(s) > 1 {
		s = s[1 : len(s)-1]
	}
	return s
}
