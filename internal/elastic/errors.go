package elastic

import (
	"strings"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// IsRecordNotFound -
func IsRecordNotFound(err error) bool {
	_, ok := err.(*RecordNotFoundError)
	return ok
}

// RecordNotFoundError -
type RecordNotFoundError struct {
	index string
	id    string
	query base
}

// NewRecordNotFoundError -
func NewRecordNotFoundError(index, id string, query base) *RecordNotFoundError {
	return &RecordNotFoundError{index, id, query}
}

// NewRecordNotFoundErrorFromResponse -
func NewRecordNotFoundErrorFromResponse(resp *esapi.Response) *RecordNotFoundError {
	return &RecordNotFoundError{resp.String(), "", nil}
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
	if e.query != nil {
		builder.WriteString("query=")
		b, _ := json.MarshalIndent(e.query, "", " ")
		builder.Write(b)
	}
	return builder.String()
}
