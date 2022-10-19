package ast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/encoding"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/pkg/errors"
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
	xi, ok := x.Value.(*types.BigInt)
	if !ok {
		return -1
	}
	yi, ok := y.Value.(*types.BigInt)
	if !ok {
		return 1
	}
	return xi.Cmp(yi.Int)
}

func compareAddresses(x, y *Address) (int, error) {
	if x == nil || y == nil {
		return 0, errors.Errorf("invalid comparable addresses: %v and %v", x, y)
	}
	if x.ValueKind != valueKindString {
		value, err := forge.UnforgeAddress(x.Value.(string))
		if err != nil {
			return 0, err
		}
		x.ValueKind = valueKindString
		x.Value = value
	}

	if y.ValueKind != valueKindString {
		value, err := forge.UnforgeAddress(y.Value.(string))
		if err != nil {
			return 0, err
		}
		y.ValueKind = valueKindString
		y.Value = value
	}

	xIsContract := strings.HasPrefix(x.Value.(string), encoding.PrefixPublicKeyKT1)
	yIsContract := strings.HasPrefix(y.Value.(string), encoding.PrefixPublicKeyKT1)

	switch {
	case xIsContract && !yIsContract:
		return 1, nil
	case !xIsContract && yIsContract:
		return -1, nil
	default:
		return strings.Compare(x.Value.(string), y.Value.(string)), nil
	}
}
