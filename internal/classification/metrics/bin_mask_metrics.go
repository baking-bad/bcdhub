package metrics

import (
	"math/bits"
	"reflect"
	"strings"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/pkg/errors"
)

// BinMask -
type BinMask struct {
	Field string
}

// NewBinMask -
func NewBinMask(field string) *BinMask {
	return &BinMask{
		Field: field,
	}
}

// Compute -
func (m *BinMask) Compute(a, b contract.Contract) Feature {
	f := Feature{
		Name: strings.ToLower(m.Field),
	}

	mask1, err := m.getContractFieldBinMask(a)
	if err != nil {
		logger.Err(err)
		return f
	}

	mask2, err := m.getContractFieldBinMask(b)
	if err != nil {
		logger.Err(err)
		return f
	}

	if mask1 == mask2 {
		f.Value = 1
	} else {
		rate := float64(bits.OnesCount64(uint64(mask1^mask2))) / 15.0
		f.Value = round(rate)
	}
	return f
}

func (m *BinMask) getContractFieldBinMask(c contract.Contract) (int64, error) {
	r := reflect.ValueOf(c)
	f := reflect.Indirect(r).FieldByName(m.Field)

	switch f.Kind() {
	case reflect.Int64:
		return f.Int(), nil
	default:
		return -1, errors.Errorf("Invalid field %s type: %v", m.Field, f.Kind())
	}
}
