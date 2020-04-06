package cerrors

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/contractparser/unpack"
	"github.com/baking-bad/bcdhub/internal/contractparser/unpack/rawbytes"
	"github.com/tidwall/gjson"
)

// IError -
type IError interface {
	Parse(data gjson.Result)
	Is(string) bool
	Format() error
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

// Format -
func (e *DefaultError) Format() error {
	if e.With == "" {
		return nil
	}
	text := gjson.Parse(e.With)
	if text.Get("bytes").Exists() {
		data := text.Get("bytes").String()
		data = strings.TrimPrefix(data, unpack.MainPrefix)
		decodedString, err := rawbytes.ToMicheline(data)
		if err == nil {
			text = gjson.Parse(decodedString)
		}
	}
	errString, err := formatter.MichelineToMichelson(text, true, formatter.DefLineSize)
	if err != nil {
		return err
	}
	e.With = errString
	return nil
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
	id := data.Get("id").String()
	var e IError
	if strings.Contains(id, balanceTooLow) {
		e = &BalanceTooLowError{}
	} else {
		e = &DefaultError{}
	}
	e.Parse(data)
	return e
}
