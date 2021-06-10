package metrics

import (
	"reflect"
	"strings"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/pkg/errors"
)

// Array -
type Array struct {
	Field string
}

// NewArray -
func NewArray(field string) *Array {
	return &Array{
		Field: field,
	}
}

// Compute -
func (m *Array) Compute(a, b contract.Contract) Feature {
	f := Feature{
		Name: strings.ToLower(m.Field),
	}

	aArr, err := m.getContractFieldArray(a)
	if err != nil {
		logger.Error(err)
		return f
	}

	bArr, err := m.getContractFieldArray(b)
	if err != nil {
		logger.Error(err)
		return f
	}

	sum := 0.0
	if len(aArr) == 0 && len(bArr) == 0 {
		f.Value = 1
		return f
	} else if len(aArr) == 0 || len(bArr) == 0 {
		return f
	}

	for i := range aArr {
		found := false

		for j := range bArr {
			if bArr[j] == aArr[i] {
				found = true
				break
			}
		}

		if found {
			sum += 2
		}
	}

	f.Value = round(sum / float64(len(aArr)+len(bArr)))
	return f
}

func (m *Array) getContractFieldArray(c contract.Contract) ([]interface{}, error) {
	r := reflect.ValueOf(c)
	f := reflect.Indirect(r).FieldByName(m.Field)

	switch f.Kind() {
	case reflect.Slice, reflect.Array:
		ret := make([]interface{}, f.Len())
		for i := 0; i < f.Len(); i++ {
			ret[i] = f.Index(i).Interface()
		}
		return ret, nil
	default:
		return nil, errors.Errorf("Invalid field %s type: %v", m.Field, f.Kind())
	}
}
