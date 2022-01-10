package storage

import (
	"encoding/hex"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	defaultTimeout = time.Second * 10
)

// DecodeValue -
func DecodeValue(value []byte) string {
	var tree ast.UntypedAST
	if err := json.Unmarshal(value, &tree); err != nil {
		return ""
	}
	s, err := tree.GetStrings(true)
	if err != nil || len(s) == 0 {
		return ""
	}
	return sanitizeString(s[0])
}

func sanitizeString(s string) string {
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") && len(s) > 1 {
		s = s[1 : len(s)-1]
	}
	return s
}

func decodeData(value []byte) ([]byte, error) {
	jsonValue := gjson.ParseBytes(value)
	if !jsonValue.Get("bytes").Exists() {
		return nil, nil
	}
	return hex.DecodeString(jsonValue.Get("bytes").String())
}
