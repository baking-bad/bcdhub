package search

import "regexp"

var ptrRegEx = regexp.MustCompile(`^ptr:\d+$`)

// IsPtrSearch - check searchString on `ptr:%d` pattern
func IsPtrSearch(searchString string) bool {
	return ptrRegEx.MatchString(searchString)
}
