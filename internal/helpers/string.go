package helpers

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/ipfs/go-cid"
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

// SpaceStringsBuilder -
func SpaceStringsBuilder(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

// Slug -
func Slug(alias string) string {
	return strings.ReplaceAll(strings.ToLower(alias), " ", "-")
}

// IsIPFS -
func IsIPFS(hash string) bool {
	_, err := cid.Decode(hash)
	return err != nil
}

// Escape -
func Escape(str string) string {
	return strings.ReplaceAll(str, "\u0000", `\u0000`)
}
