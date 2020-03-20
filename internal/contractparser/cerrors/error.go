package cerrors

import (
	"strings"

	"github.com/tidwall/gjson"
)

// Error -
type Error struct {
	Title       string `json:"title"`
	Description string `json:"descr"`
	Kind        string `json:"kind"`
	Location    int64  `json:"location,omitempty"`
	With        string `json:"with,omitempty"`
	ID          string `json:"id"`
}

// Parse - parse error from json
func (e *Error) Parse(data gjson.Result) {
	e.ID = data.Get("id").String()
	e.Kind = data.Get("kind").String()
	e.Location = data.Get("location").Int()
	e.With = data.Get("with").String()

	if !errorDescriptions.IsObject() {
		return
	}

	parts := strings.Split(e.ID, ".")
	if len(parts) > 1 {
		parts = parts[2:]
	}
	errorID := strings.Join(parts, ".")
	errorID = strings.Replace(errorID, ".", "\\.", -1)
	errDescr := errorDescriptions.Get(errorID)
	if errDescr.Exists() {
		e.Title = errDescr.Get("title").String()
		e.Description = errDescr.Get("descr").String()
	}
}

// ParseArray -
func ParseArray(data gjson.Result) []Error {
	if !data.IsArray() {
		return nil
	}

	ret := make([]Error, 0)
	for _, item := range data.Array() {
		var e Error
		e.Parse(item)
		ret = append(ret, e)
	}
	return ret
}
