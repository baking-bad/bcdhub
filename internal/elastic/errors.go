package elastic

import "strings"

// IsRecordNotFound -
func IsRecordNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), RecordNotFound)
}
