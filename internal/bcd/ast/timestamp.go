package ast

import (
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/pkg/errors"
)

//
//  Timestamp
//

// Bounds of timestamps representable in RFC3339: 0001-01-01T00:00:00Z and 9999-12-31T23:59:59Z.
// Timestamps outside this range are rendered as integers, the same way the Tezos node does.
var (
	minRFC3339Timestamp = big.NewInt(-62135596800)
	maxRFC3339Timestamp = big.NewInt(253402300799)
)

// Timestamp - Michelson timestamp. Value is always *types.BigInt with seconds since Unix epoch (or nil if unset).
type Timestamp struct {
	Default
}

// NewTimestamp -
func NewTimestamp(depth int) *Timestamp {
	return &Timestamp{
		Default: NewDefault(consts.TIMESTAMP, 0, depth),
	}
}

// ParseValue -
func (t *Timestamp) ParseValue(node *base.Node) error {
	switch {
	case node.IntValue != nil:
		if node.IntValue.Int == nil {
			return errors.Wrap(consts.ErrInvalidType, "Timestamp.ParseValue: empty integer")
		}
		t.setSeconds(node.IntValue)
	case node.StringValue != nil:
		if *node.StringValue == "" {
			t.setSeconds(types.NewBigInt(0))
			return nil
		}
		if value, err := parseTimestampString(*node.StringValue); err == nil {
			t.setSeconds(value)
		}
	default:
		return errors.Wrap(consts.ErrInvalidType, "Timestamp.ParseValue: expected int or string")
	}
	return nil
}

// ToBaseNode -
func (t *Timestamp) ToBaseNode(optimized bool) (*base.Node, error) {
	seconds, err := t.seconds()
	if err != nil {
		return nil, err
	}
	if !optimized {
		if formatted, ok := formatTimestamp(seconds); ok {
			return toBaseNodeString(formatted), nil
		}
	}
	return toBaseNodeInt(seconds), nil
}

// ToMiguel -
func (t *Timestamp) ToMiguel() (*MiguelNode, error) {
	name := t.GetTypeName()
	return &MiguelNode{
		Prim:  t.Prim,
		Type:  t.Prim,
		Name:  &name,
		Value: t.displayValue(),
	}, nil
}

// FromJSONSchema -
func (t *Timestamp) FromJSONSchema(data map[string]interface{}) error {
	key := t.GetName()
	value, ok := data[key]
	if !ok {
		return nil
	}

	switch val := value.(type) {
	case nil:
		// explicit null is kept as an empty integer, consistently with int/nat/mutez
		t.setSeconds(&types.BigInt{})
	case string:
		seconds, err := parseTimestampString(val)
		if err != nil {
			return errors.Wrapf(consts.ErrValidation, "time should be in RFC3339 or unix seconds %s=%s", key, val)
		}
		t.setSeconds(seconds)
	case float64:
		if math.IsInf(val, 0) || math.IsNaN(val) || val != math.Trunc(val) {
			return errors.Wrapf(consts.ErrValidation, "timestamp should be an integer number of seconds %s=%v", key, val)
		}
		seconds, _ := big.NewFloat(val).Int(nil)
		t.setSeconds(&types.BigInt{Int: seconds})
	default:
		return errors.Wrapf(consts.ErrValidation, "expected a timestamp value: %s=%v", key, value)
	}
	return nil
}

// ToParameters -
func (t *Timestamp) ToParameters() ([]byte, error) {
	value, ok := t.Value.(*types.BigInt)
	if !ok || value == nil {
		return nil, errors.Wrapf(consts.ErrValidation, "missing value for parameter %s", t.GetName())
	}
	seconds, err := value.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return fmt.Appendf(nil, `{"int":%s}`, seconds), nil
}

// ToJSONSchema -
func (t *Timestamp) ToJSONSchema() (*JSONSchema, error) {
	return wrapObject(&JSONSchema{
		Prim:    t.Prim,
		Title:   t.GetName(),
		Type:    JSONSchemaTypeString,
		Format:  "date-time",
		Default: time.Now().UTC().Format(time.RFC3339),
	}), nil
}

// Compare -
func (t *Timestamp) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*Timestamp)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	x, err := t.seconds()
	if err != nil {
		return 0, err
	}
	y, err := secondItem.seconds()
	if err != nil {
		return 0, err
	}
	return x.Cmp(y.Int), nil
}

// Distinguish -
func (t *Timestamp) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Timestamp)
	if !ok {
		return nil, nil
	}
	node, err := t.Default.Distinguish(&second.Default)
	if err != nil || node == nil {
		return node, err
	}
	node.Value = t.displayValue()
	if node.From != nil {
		node.From = second.displayValue()
	}
	return node, nil
}

// GetJSONModel -
func (t *Timestamp) GetJSONModel(model JSONModel) {
	if model == nil {
		return
	}
	value := t.displayValue()
	if value == nil {
		value = ""
	}
	model[t.GetName()] = value
}

// FindByName -
func (t *Timestamp) FindByName(name string, isEntrypoint bool) Node {
	if t.GetName() == name {
		return t
	}
	return nil
}

func (t *Timestamp) setSeconds(value *types.BigInt) {
	t.Value = value
	t.ValueKind = valueKindInt
}

func (t *Timestamp) seconds() (*types.BigInt, error) {
	value, ok := t.Value.(*types.BigInt)
	if !ok || value == nil || value.Int == nil {
		return nil, errors.Wrapf(consts.ErrValidation, "missing value for parameter %s", t.GetName())
	}
	return value, nil
}

// displayValue - RFC3339 string if timestamp fits into it, decimal seconds otherwise. Nil if value is unset.
func (t *Timestamp) displayValue() interface{} {
	seconds, err := t.seconds()
	if err != nil {
		return nil
	}
	if formatted, ok := formatTimestamp(seconds); ok {
		return formatted
	}
	return seconds.String()
}

func formatTimestamp(seconds *types.BigInt) (string, bool) {
	if seconds.Cmp(minRFC3339Timestamp) < 0 || seconds.Cmp(maxRFC3339Timestamp) > 0 {
		return "", false
	}
	return time.Unix(seconds.Int64(), 0).UTC().Format(time.RFC3339), true
}

// parseTimestampString - accepts RFC3339 strings and decimal unix seconds.
func parseTimestampString(value string) (*types.BigInt, error) {
	value = strings.TrimSpace(value)
	if ts, err := time.Parse(time.RFC3339, normalizeRFC3339(value)); err == nil {
		return types.NewBigInt(ts.Unix()), nil
	}
	seconds, err := types.NewBigIntFromString(value)
	if err != nil {
		return nil, errors.Wrapf(consts.ErrInvalidType, "invalid timestamp: %q", value)
	}
	return seconds, nil
}

// normalizeRFC3339 - Tezos parses timestamps leniently and also accepts a space or a lowercase 't'
// as the date-time separator and a lowercase 'z' as UTC offset, which time.RFC3339 rejects.
func normalizeRFC3339(value string) string {
	if value == "" {
		return value
	}
	b := []byte(value)
	if len(b) > 10 && (b[10] == ' ' || b[10] == 't') {
		b[10] = 'T'
	}
	if b[len(b)-1] == 'z' {
		b[len(b)-1] = 'Z'
	}
	return string(b)
}
