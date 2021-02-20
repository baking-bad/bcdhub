package tezerrors

import (
	"bytes"
	"fmt"
	"strings"

	stdJSON "encoding/json"

	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// ParseArray -
func ParseArray(data []byte) ([]*Error, error) {
	if len(data) == 0 {
		return nil, nil
	}
	ret := make([]*Error, 0)
	err := json.Unmarshal(data, &ret)
	return ret, err
}

func getErrorID(id string) string {
	parts := strings.Split(id, ".")
	if len(parts) > 1 {
		parts = parts[2:]
	}
	return strings.Join(parts, ".")
}

// IError -
type IError interface {
	fmt.Stringer

	Format() error
}

// Error -
type Error struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Title       string `json:"title,omitempty"`
	Description string `json:"descr,omitempty"`

	IError `json:"-"`
}

// GetTitle -
func (e *Error) GetTitle() string {
	return e.Title
}

// Is -
func (e *Error) Is(errorID string) bool {
	return strings.Contains(e.ID, errorID)
}

// Format -
func (e *Error) Format() error {
	if e.IError == nil {
		return nil
	}
	return e.IError.Format()
}

// String -
func (e *Error) String() string {
	return e.ID
}

// UnmarshalJSON -
func (e *Error) UnmarshalJSON(data []byte) error {
	var typ struct {
		ID   string `json:"id"`
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(data, &typ); err != nil {
		return err
	}
	e.ID = typ.ID
	e.Kind = typ.Kind

	errorID := getErrorID(e.ID)
	var descr Description
	var ok bool
	if descr, ok = errorDescriptions[errorID]; !ok {
		if err := json.Unmarshal(data, &descr); err != nil {
			return err
		}
	}
	e.Title = descr.Title
	e.Description = descr.Description

	switch {
	case strings.Contains(e.ID, consts.BalanceTooLowError):
		e.IError = new(BalanceTooLowError)
	case strings.Contains(e.ID, consts.InvalidSyntacticConstantError):
		e.IError = new(InvalidSyntacticConstantError)
	default:
		e.IError = new(DefaultError)
	}
	return json.Unmarshal(data, e.IError)
}

// MarshalJSON -
func (e *Error) MarshalJSON() ([]byte, error) {
	type eBuf *Error
	data, err := json.Marshal(eBuf(e))
	if err != nil {
		return nil, err
	}
	data = data[:len(data)-1]
	w := bytes.NewBuffer(data)

	body, err := json.Marshal(e.IError)
	if err != nil {
		return nil, err
	}
	if len(body) > 2 {
		body = body[1:]

		w.WriteString(", ")
		w.Write(body)
	} else {
		w.WriteByte('}')
	}

	return w.Bytes(), nil
}

// DefaultError -
type DefaultError struct {
	Location int64              `json:"location,omitempty"`
	With     stdJSON.RawMessage `json:"with,omitempty"`
}

// Format -
func (e *DefaultError) Format() error {
	if e.With == nil {
		return nil
	}
	var tree ast.UntypedAST
	if err := json.Unmarshal(e.With, &tree); err != nil {
		return err
	}

	if len(tree) == 0 {
		return nil
	}

	text := string(e.With)
	if tree[0].BytesValue != nil {
		subTree, err := forge.UnpackString(*tree[0].BytesValue)
		if err == nil {
			text, _ = json.MarshalToString(subTree)
		}
	}
	if text != "" {
		errString, err := formatter.MichelineStringToMichelson(text, true, formatter.DefLineSize)
		if err != nil {
			return err
		}
		e.With = []byte(errString)
	}
	return nil
}

// String -
func (e *DefaultError) String() string {
	return string(e.With)
}

// BalanceTooLowError -
type BalanceTooLowError struct {
	DefaultError
	Balance int64 `json:"balance,string"`
	Amount  int64 `json:"amount,string"`
}

// Format -
func (e *BalanceTooLowError) Format() error {
	return nil
}

// String -
func (e *BalanceTooLowError) String() string {
	return fmt.Sprintf("Balance too low: %d < %d", e.Balance, e.Amount)
}

// InvalidSyntacticConstantError -
type InvalidSyntacticConstantError struct {
	WrongExpressionSnake stdJSON.RawMessage `json:"wrong_expression"`
	ExpectedFormSnake    stdJSON.RawMessage `json:"expected_form"`

	WrongExpressionCamel stdJSON.RawMessage `json:"wrongExpression"`
	ExpectedFormCamel    stdJSON.RawMessage `json:"expectedForm"`
}

func (e *InvalidSyntacticConstantError) getWrongExpression() []byte {
	wrongExpr := e.ExpectedFormCamel
	if wrongExpr == nil {
		wrongExpr = e.WrongExpressionCamel
	}
	return wrongExpr
}

func (e *InvalidSyntacticConstantError) getExpectedForm() []byte {
	expForm := e.ExpectedFormCamel
	if expForm == nil {
		expForm = e.ExpectedFormSnake
	}
	return expForm
}

// String -
func (e *InvalidSyntacticConstantError) String() string {
	return string(e.getWrongExpression())
}

// Format -
func (e *InvalidSyntacticConstantError) Format() error {
	wrongExpr := e.getWrongExpression()
	if wrongExpr != nil {
		wrongExpression, err := formatter.MichelineToMichelson(gjson.ParseBytes(wrongExpr), false, formatter.DefLineSize)
		if err != nil {
			return err
		}
		e.WrongExpressionCamel = []byte(wrongExpression)
	}

	expForm := e.getExpectedForm()
	if expForm != nil {
		expectedForm, err := formatter.MichelineToMichelson(gjson.ParseBytes(expForm), false, formatter.DefLineSize)
		if err != nil {
			return err
		}
		e.ExpectedFormCamel = []byte(expectedForm)
	}
	return nil
}
