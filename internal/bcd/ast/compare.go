package ast

import (
	"reflect"
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

// distinguishReadable - Distinguish for types whose value is stored either in optimized (bytes)
// or in readable (string) form. Both values are compared in readable form, the nodes are not modified.
func distinguishReadable(x, y *Default, unforge func(string) (string, error)) (*MiguelNode, error) {
	readableX, err := readableDefault(x, unforge)
	if err != nil {
		return nil, err
	}
	readableY, err := readableDefault(y, unforge)
	if err != nil {
		return nil, err
	}
	return readableX.Distinguish(readableY)
}

// readableDefault - returns a copy of `d` with the value converted to readable form
func readableDefault(d *Default, unforge func(string) (string, error)) (*Default, error) {
	readable := *d
	if readable.ValueKind != valueKindBytes {
		return &readable, nil
	}
	value, ok := readable.Value.(string)
	if !ok {
		return &readable, nil
	}
	decoded, err := unforge(value)
	if err != nil {
		return nil, err
	}
	readable.Value = decoded
	readable.ValueKind = valueKindString
	return &readable, nil
}

// valuesEqual - compares values of scalar nodes. Unlike `==` it doesn't panic on mixed or uncomparable types.
func valuesEqual(x, y interface{}) bool {
	xi, xIsInt := x.(*types.BigInt)
	yi, yIsInt := y.(*types.BigInt)
	switch {
	case xIsInt && yIsInt:
		xEmpty := xi == nil || xi.Int == nil
		yEmpty := yi == nil || yi.Int == nil
		if xEmpty || yEmpty {
			return xEmpty == yEmpty
		}
		return xi.Cmp(yi.Int) == 0
	case xIsInt || yIsInt:
		return false
	default:
		return reflect.DeepEqual(x, y)
	}
}

// distinguishNodes - `current.Distinguish(prev)` for containers. Distinguish returns no node when it can't
// compare the values (different types or unset values): then the current value is returned without diff.
// The result is still nil when the current node has no representation (e.g. empty big_map).
func distinguishNodes(current, prev Node) (*MiguelNode, error) {
	node, err := current.Distinguish(prev)
	if err != nil || node != nil {
		return node, err
	}
	return current.ToMiguel()
}
