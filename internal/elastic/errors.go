package elastic

import (
	"strings"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// IsRecordNotFound -
func (e *Elastic) IsRecordNotFound(err error) bool {
	_, ok := err.(*RecordNotFoundError)
	return ok
}

// RecordNotFoundError -
type RecordNotFoundError struct {
	index string
	id    string
}

// NewRecordNotFoundError -
func NewRecordNotFoundError(index, id string) *RecordNotFoundError {
	return &RecordNotFoundError{index, id}
}

// NewRecordNotFoundErrorFromResponse -
func NewRecordNotFoundErrorFromResponse(resp *esapi.Response) *RecordNotFoundError {
	return &RecordNotFoundError{resp.String(), ""}
}

// Error -
func (e *RecordNotFoundError) Error() string {
	var builder strings.Builder
	builder.WriteString("Record is not found: ")
	if e.index != "" {
		builder.WriteString("index=")
		builder.WriteString(e.index)
		builder.WriteString(" ")
	}
	if e.id != "" {
		builder.WriteString("id=")
		builder.WriteString(e.id)
		builder.WriteString(" ")
	}
	return builder.String()
}
