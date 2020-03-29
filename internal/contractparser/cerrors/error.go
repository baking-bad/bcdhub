package cerrors

import (
	"strings"

	"github.com/tidwall/gjson"
)

// IError -
type IError interface {
	Parse(data gjson.Result)
	Is(string) bool
}

// DefaultError -
type DefaultError struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"descr"`
	Kind        string `json:"kind"`
	Location    int64  `json:"location,omitempty"`
	With        string `json:"with,omitempty"`
}

// Parse - parse error from json
func (e *DefaultError) Parse(data gjson.Result) {
	e.ID = data.Get("id").String()
	e.Kind = data.Get("kind").String()
	if !errorDescriptions.IsObject() {
		return
	}
	errorID := getErrorID(data)
	errDescr := errorDescriptions.Get(errorID)
	if errDescr.Exists() {
		e.Title = errDescr.Get("title").String()
		e.Description = errDescr.Get("descr").String()
	}
	e.Location = data.Get("location").Int()
	e.With = data.Get("with").String()
}

// Is -
func (e *DefaultError) Is(errorID string) bool {
	return strings.Contains(e.ID, errorID)
}

// BalanceTooLowError -
type BalanceTooLowError struct {
	DefaultError

	Balance int64 `json:"balance"`
	Amount  int64 `json:"amount"`
}

// Parse - parse error from json
func (e *BalanceTooLowError) Parse(data gjson.Result) {
	e.DefaultError = DefaultError{}
	e.DefaultError.Parse(data)

	e.Balance = data.Get("balance").Int()
	e.Amount = data.Get("amount").Int()
}

// ParseArray -
func ParseArray(data gjson.Result) []IError {
	if !data.IsArray() {
		return nil
	}

	ret := make([]IError, 0)
	for _, item := range data.Array() {
		ret = append(ret, getErrorObject(item))
	}
	return ret
}

func getErrorID(data gjson.Result) string {
	id := data.Get("id").String()
	parts := strings.Split(id, ".")
	if len(parts) > 1 {
		parts = parts[2:]
	}
	errorID := strings.Join(parts, ".")
	return strings.Replace(errorID, ".", "\\.", -1)
}

func getErrorObject(data gjson.Result) IError {
	id := getErrorID(data)
	var e IError
	switch id {
	case balanceTooLow:
		e = &BalanceTooLowError{}
	default:
		e = &DefaultError{}
	}
	e.Parse(data)
	return e
}
