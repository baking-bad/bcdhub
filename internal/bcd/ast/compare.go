package ast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/types"
)

func compareNotOptimizedTypes(x, y Default, optimizer func(string) (string, error)) (int, error) {
	if x.ValueKind != valueKindBytes {
		value, err := optimizer(x.Value.(string))
		if err != nil {
			return 0, err
		}
		x.ValueKind = valueKindBytes
		x.Value = value
	}
	if y.ValueKind != valueKindBytes {
		value, err := optimizer(y.Value.(string))
		if err != nil {
			return 0, err
		}
		y.ValueKind = valueKindBytes
		y.Value = value
	}

	return strings.Compare(x.Value.(string), y.Value.(string)), nil
}

func compareBigInt(x, y Default) int {
	xi := x.Value.(*types.BigInt)
	yi := y.Value.(*types.BigInt)
	return xi.Cmp(yi.Int)
}
