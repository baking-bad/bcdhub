package ast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/pkg/errors"
)

// Set -
type Set struct {
	Default

	Type Node
	Data []Node
}

// NewSet -
func NewSet(depth int) *Set {
	return &Set{
		Default: NewDefault(consts.SET, -1, depth),
	}
}

// String -
func (set *Set) String() string {
	var s strings.Builder

	s.WriteString(set.Default.String())
	if len(set.Data) > 0 {
		for i := range set.Data {
			s.WriteString(strings.Repeat(base.DefaultIndent, set.depth))
			s.WriteByte('{')
			s.WriteByte('\n')
			s.WriteString(strings.Repeat(base.DefaultIndent, set.depth+1))
			s.WriteString(set.Data[i].String())
			s.WriteString(strings.Repeat(base.DefaultIndent, set.depth))
			s.WriteByte('}')
			s.WriteByte('\n')
		}
	} else {
		s.WriteString(strings.Repeat(base.DefaultIndent, set.depth))
		s.WriteString(set.Type.String())
	}
	return s.String()
}

// MarshalJSON -
func (set *Set) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.SET, set.annots, set.Type)
}

// ParseType -
func (set *Set) ParseType(node *base.Node, id *int) error {
	if err := set.Default.ParseType(node, id); err != nil {
		return err
	}

	typ, err := typingNode(node.Args[0], set.depth, id)
	if err != nil {
		return err
	}
	set.Type = typ

	return nil
}

// ParseValue -
func (set *Set) ParseValue(node *base.Node) error {
	if node.Prim != base.PrimArray {
		return errors.Wrap(base.ErrInvalidPrim, "List.ParseValue")
	}

	set.Data = make([]Node, 0)

	for i := range node.Args {
		item, err := createByType(set.Type)
		if err != nil {
			return err
		}
		if err := item.ParseValue(node.Args[i]); err != nil {
			return err
		}
		set.Data = append(set.Data, item)
	}

	return nil
}

// ToMiguel -
func (set *Set) ToMiguel() (*MiguelNode, error) {
	node, err := set.Default.ToMiguel()
	if err != nil {
		return nil, err
	}

	node.Children = make([]*MiguelNode, 0)
	for i := range set.Data {
		child, err := set.Data[i].ToMiguel()
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	}

	return node, nil
}

// ToBaseNode -
func (set *Set) ToBaseNode(optimized bool) (*base.Node, error) {
	return arrayToBaseNode(set.Data, optimized)
}

// ToJSONSchema -
func (set *Set) ToJSONSchema() (*JSONSchema, error) {
	s := &JSONSchema{
		Prim:    set.Prim,
		Type:    JSONSchemaTypeArray,
		Title:   set.GetName(),
		Default: make([]interface{}, 0),
		Items: &SchemaKey{
			Type:       JSONSchemaTypeObject,
			Required:   make([]string, 0),
			Properties: make(map[string]*JSONSchema),
		},
	}

	if err := setChildSchema(set.Type, true, s); err != nil {
		return nil, err
	}

	return &JSONSchema{
		Type: JSONSchemaTypeObject,
		Properties: map[string]*JSONSchema{
			set.GetName(): s,
		},
	}, nil
}

// FromJSONSchema -
func (set *Set) FromJSONSchema(data map[string]interface{}) error {
	var arr []interface{}
	for key := range data {
		if key == set.GetName() {
			val := data[key]
			arrVal, ok := val.([]interface{})
			if !ok {
				return errors.Wrapf(base.ErrInvalidType, "Set.FromJSONSchema %T", val)
			}
			arr = arrVal
			break
		}
	}
	if arr == nil {
		return errors.Wrap(base.ErrJSONDataIsAbsent, "Set.FromJSONSchema")
	}

	for i := range arr {
		item, ok := arr[i].(map[string]interface{})
		if !ok {
			return errors.Wrap(base.ErrValidation, "Set.FromJSONSchema")
		}
		itemTree, err := createByType(set.Type)
		if err != nil {
			return err
		}
		for key := range item {
			itemMap := item[key].(map[string]interface{})
			if err := itemTree.FromJSONSchema(itemMap); err != nil {
				return err
			}
		}
		set.Data = append(set.Data, itemTree)
	}
	return nil
}

// EnrichBigMap -
func (set *Set) EnrichBigMap(bmd []*base.BigMapDiff) error {
	for i := range set.Data {
		if err := set.Data[i].EnrichBigMap(bmd); err != nil {
			return err
		}
	}
	return nil
}

// ToParameters -
func (set *Set) ToParameters() ([]byte, error) {
	return buildListParameters(set.Data)
}
