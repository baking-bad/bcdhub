package ast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
)

// TxRollupL2Address -
type TxRollupL2Address Address

// NewTxRollupL2Address -
func NewTxRollupL2Address(depth int) *TxRollupL2Address {
	return &TxRollupL2Address{
		Default: NewDefault(consts.TXROLLUPL2ADDRESS, 0, depth),
	}
}

// Compare -
func (a *TxRollupL2Address) Compare(second Comparable) (int, error) {
	secondAddress, ok := second.(*TxRollupL2Address)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	if a.Value == secondAddress.Value {
		return 0, nil
	}
	if a.ValueKind == secondAddress.ValueKind {
		return strings.Compare(a.Value.(string), secondAddress.Value.(string)), nil
	}
	return compareNotOptimizedTypes(a.Default, secondAddress.Default, forge.Contract)
}

// Distinguish -
func (a *TxRollupL2Address) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*TxRollupL2Address)
	if !ok {
		return nil, nil
	}
	if err := a.optimizeStringValue(forge.UnforgeContract); err != nil {
		return nil, err
	}
	if err := second.optimizeStringValue(forge.UnforgeContract); err != nil {
		return nil, err
	}
	return a.Default.Distinguish(&second.Default)
}
