package ast

import (
	"fmt"
	"strconv"
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
		i := node.IntValue.Int64()
		if 253402300799 > i { // 31 December 9999 23:59:59 Golang time restriction
			t.Value = time.Unix(i, 0).UTC()
		} else {
			t.Value = fmt.Sprintf("%d", i)
		}
	case node.StringValue != nil:
		utc, err := time.Parse(time.RFC3339, *node.StringValue)
		if err != nil {
			if *node.StringValue == "" {
				t.Value = time.Unix(0, 0).UTC()
			} else {
				i, err := strconv.ParseInt(*node.StringValue, 10, 64)
				if err != nil {
					return err
				}
				if 253402300799 > i { // 31 December 9999 23:59:59 Golang time restriction
					t.Value = time.Unix(i, 0).UTC()
				} else {
					t.Value = fmt.Sprintf("%d", i)
				}
			}
		} else {
			t.Value = utc.UTC()
		}
	}
	return nil
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
	for key := range data {
		if key == t.GetName() {
			t.ValueKind = valueKindInt
			switch val := data[key].(type) {
			case string:
				ts, err := time.Parse(time.RFC3339, val)
				if err != nil {
					return err
				}
				t.Value = types.NewBigInt(ts.UTC().Unix())
			case float64:
				t.Value = types.NewBigInt(int64(val))
			}
			break
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
