package elastic

import "strings"

// IsRecordNotFound -
func IsRecordNotFound(err error) bool {
	return strings.Contains(err.Error(), RecordNotFound)
}
