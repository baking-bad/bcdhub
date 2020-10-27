package cerrors

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
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
	String() string
	GetTitle() string
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
		e.Title = data.Get("title").String()
		e.Description = data.Get("descr").String()
	} else {
		errorID := getErrorID(data)
		errDescr := errorDescriptions.Get(errorID)
		if errDescr.Exists() {
			e.Title = errDescr.Get("title").String()
			e.Description = errDescr.Get("descr").String()
		}
	}
	e.Location = data.Get("location").Int()
	e.With = data.Get("with").String()
}

// Is -
func (e *DefaultError) Is(errorID string) bool {
	return strings.Contains(e.ID, errorID)
}

// GetTitle -
func (e *DefaultError) GetTitle() string {
	return e.Title
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

// String -
func (e *DefaultError) String() string {
	return e.With
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

	switch {
	case strings.Contains(id, consts.BalanceTooLowError):
		e = &BalanceTooLowError{}
	case strings.Contains(id, consts.InvalidSyntacticConstantError):
		e = &InvalidSyntacticConstantError{}
	default:
		e = &DefaultError{}
	}

	e.Parse(data)
	return e
}

// InvalidSyntacticConstantError -
type InvalidSyntacticConstantError struct {
	DefaultError

	WrongExpression string `json:"wrong_expression"`
	ExpectedForm    string `json:"expected_form"`
}

// Parse - parse error from json
func (e *InvalidSyntacticConstantError) Parse(data gjson.Result) {
	e.DefaultError = DefaultError{}
	e.DefaultError.Parse(data)

	for _, key := range []string{"wrongExpression", "wrong_expression"} {
		item := data.Get(key)
		if item.Exists() {
			e.WrongExpression = item.String()
			break
		}
	}

	for _, key := range []string{"expectedForm", "expected_form"} {
		item := data.Get(key)
		if item.Exists() {
			e.ExpectedForm = item.String()
			break
		}
	}
}

// String -
func (e *InvalidSyntacticConstantError) String() string {
	return gjson.Parse(e.WrongExpression).Raw
}

// Format -
func (e *InvalidSyntacticConstantError) Format() error {
	if err := e.DefaultError.Format(); err != nil {
		return err
	}
	if e.WrongExpression != "" {
		wrongExpression, err := formatter.MichelineToMichelson(gjson.Parse(e.WrongExpression), false, formatter.DefLineSize)
		if err != nil {
			return err
		}
		e.WrongExpression = wrongExpression
	}

	if e.ExpectedForm != "" {
		expectedForm, err := formatter.MichelineToMichelson(gjson.Parse(e.ExpectedForm), false, formatter.DefLineSize)
		if err != nil {
			return err
		}
		e.ExpectedForm = expectedForm
	}
	return nil
}
