package metrics

import (
	"reflect"
	"strings"

	"github.com/baking-bad/bcdhub/internal/models/contract"
)

// Bool -
type Bool struct {
	Field string
}

// NewBool -
func NewBool(field string) *Bool {
	return &Bool{
		Field: field,
	}
}

// Compute -
func (m *Bool) Compute(a, b contract.Script) Feature {
	f := Feature{
		Name: strings.ToLower(m.Field),
	}
	aVal := m.getContractField(a)
	bVal := m.getContractField(b)

	if aVal == bVal {
		f.Value = 1
	}
	return f
}

func (m *Bool) getContractField(c contract.Script) interface{} {
	r := reflect.ValueOf(c)
	return reflect.Indirect(r).FieldByName(m.Field).Interface()
}
