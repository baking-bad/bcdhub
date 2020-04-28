package helpers

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/google/uuid"
)

// StringInArray -
func StringInArray(s string, arr []string) bool {
	for i := range arr {
		if arr[i] == s {
			return true
		}
	}
	return false
}

// GenerateID -
func GenerateID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// URLJoin -
func URLJoin(baseURL, queryPath string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Sprintf("%s/%s", baseURL, queryPath)
	}
	u.Path = path.Join(u.Path, queryPath)
	return u.String()
}
