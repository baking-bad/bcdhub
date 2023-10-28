package helpers

import (
	"net/url"
	"os"
	"path"
	"path/filepath"
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
func URLJoin(baseURL, queryPath string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, queryPath)
	return u.String(), nil
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
	return err == nil
}

// Escape -
func Escape(str string) string {
	return strings.ReplaceAll(str, "\u0000", `\u0000`)
}

// CleanPath makes a path safe for use with filepath.Join. This is done by not
// only cleaning the path, but also (if the path is relative) adding a leading
// '/' and cleaning it (then removing the leading '/'). This ensures that a
// path resulting from prepending another path will always resolve to lexically
// be a subdirectory of the prefixed path. This is all done lexically, so paths
// that include symlinks won't be safe as a result of using CleanPath.
func CleanPath(path string) string {
	// Deal with empty strings nicely.
	if path == "" {
		return ""
	}

	// Ensure that all paths are cleaned (especially problematic ones like
	// "/../../../../../" which can cause lots of issues).
	path = filepath.Clean(path)

	// If the path isn't absolute, we need to do more processing to fix paths
	// such as "../../../../<etc>/some/path". We also shouldn't convert absolute
	// paths to relative ones.
	if !filepath.IsAbs(path) {
		path = filepath.Clean(string(os.PathSeparator) + path)
		// This can't fail, as (by definition) all paths are relative to root.
		path, _ = filepath.Rel(string(os.PathSeparator), path)
	}

	// Clean the path again for good measure.
	return filepath.Clean(path)
}
