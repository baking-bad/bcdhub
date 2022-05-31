package ast

import (
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/pkg/errors"
)

//
//  Timestamp
//

// Timestamp -
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
		t.parseTimestamp(node.IntValue.Int64())
	case node.StringValue != nil:
		utc, err := time.Parse(time.RFC3339, *node.StringValue)
		if err != nil {
			if *node.StringValue == "" {
				t.Value = time.Unix(0, 0).UTC()
			} else if bi, ok := big.NewInt(0).SetString(*node.StringValue, 10); ok {
				t.parseTimestamp(bi.Int64())
			}
		} else {
			t.Value = utc.UTC()
		}
	}
	return nil
}

func (t *Timestamp) parseTimestamp(ts int64) {
	switch {
	case ts < 0: // integer overflow
		t.Value = time.Unix(math.MaxInt64, 0).UTC()
	case 253402300799 > ts: // 31 December 9999 23:59:59 Golang time restriction
		t.Value = time.Unix(ts, 0).UTC()
	default:
		t.Value = time.Unix(ts/1000, 0).UTC() // milliseconds
	}
}

// ToBaseNode -
func (t *Timestamp) ToBaseNode(optimized bool) (*base.Node, error) {
	switch ts := t.Value.(type) {
	case time.Time:
		if optimized {
			val := types.NewBigInt(ts.UTC().Unix())
			return toBaseNodeInt(val), nil
		}
		val := ts.UTC().Format(time.RFC3339)
		return toBaseNodeString(val), nil
	case string:
		return toBaseNodeString(ts), nil
	}
	return nil, errors.Errorf("Invalid timestamp type")
}

// FromJSONSchema -
func (t *Timestamp) FromJSONSchema(data map[string]interface{}) error {
	key := t.GetName()
	if value, ok := data[key]; ok {
		t.ValueKind = valueKindInt
		switch val := value.(type) {
		case string:
			ts, err := time.Parse(time.RFC3339, val)
			if err != nil {
				return errors.Wrapf(ErrValidation, "time should be in RFC3339  %s=%s", key, val)
			}
			t.Value = types.NewBigInt(ts.UTC().Unix())
		case float64:
			t.Value = types.NewBigInt(int64(val))
		}

	}
	return nil
}

// ToParameters -
func (t *Timestamp) ToParameters() ([]byte, error) {
	switch ts := t.Value.(type) {
	case time.Time:
		return []byte(fmt.Sprintf(`{"int":"%d"}`, ts.UTC().Unix())), nil
	case *types.BigInt:
		return []byte(fmt.Sprintf(`{"int":"%d"}`, ts.Int64())), nil
	default:
		return nil, errors.Wrapf(consts.ErrInvalidType, "Timestamp.ToParameters: %T", t.Value)
	}
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
	ts := t.Value.(time.Time)
	ts2 := secondItem.Value.(time.Time)
	switch {
	case ts.Equal(ts2):
		return 0, nil
	case ts.Before(ts2):
		return -1, nil
	default:
		return 1, nil
	}
}

// Distinguish -
func (t *Timestamp) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Timestamp)
	if !ok {
		return nil, nil
	}
	return t.Default.Distinguish(&second.Default)
}

// GetJSONModel -
func (t *Timestamp) GetJSONModel(model JSONModel) {
	if model == nil {
		return
	}
	ts := t.Value.(time.Time)
	model[t.GetName()] = ts.Format(time.RFC3339)
}

// FindByName -
func (t *Timestamp) FindByName(name string, isEntrypoint bool) Node {
	if t.GetName() == name {
		return t
	}
	return nil
}
