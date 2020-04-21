package helpers

import (
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
