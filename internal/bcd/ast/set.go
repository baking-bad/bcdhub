package ast

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
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
			s.WriteString(strings.Repeat(consts.DefaultIndent, set.Depth))
			s.WriteByte('{')
			s.WriteByte('\n')
			s.WriteString(strings.Repeat(consts.DefaultIndent, set.Depth+1))
			s.WriteString(set.Data[i].String())
			s.WriteString(strings.Repeat(consts.DefaultIndent, set.Depth))
			s.WriteByte('}')
			s.WriteByte('\n')
		}
	} else {
		s.WriteString(strings.Repeat(consts.DefaultIndent, set.Depth))
		s.WriteString(set.Type.String())
	}
	return s.String()
}

// MarshalJSON -
func (set *Set) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.SET, set.Annots, set.Type)
}

// ParseType -
func (set *Set) ParseType(node *base.Node, id *int) error {
	if err := set.Default.ParseType(node, id); err != nil {
		return err
	}

	typ, err := typingNode(node.Args[0], set.Depth, id)
	if err != nil {
		return err
	}
	set.Type = typ

	return nil
}

// ParseValue -
func (set *Set) ParseValue(node *base.Node) error {
	if node.Prim != consts.PrimArray {
		return errors.Wrap(consts.ErrInvalidPrim, "List.ParseValue")
	}

	set.Data = make([]Node, 0)

	for i := range node.Args {
		item := Copy(set.Type)
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
				return errors.Wrapf(consts.ErrInvalidType, "Set.FromJSONSchema %T", val)
			}
			arr = arrVal
			break
		}
	}
	if arr == nil {
		return errors.Wrap(consts.ErrJSONDataIsAbsent, "Set.FromJSONSchema")
	}

	for i := range arr {
		item, ok := arr[i].(map[string]interface{})
		if !ok {
			return errors.Wrap(consts.ErrValidation, "Set.FromJSONSchema")
		}
		itemTree := Copy(set.Type)
		if err := itemTree.FromJSONSchema(item); err != nil {
			return err
		}

		set.Data = append(set.Data, itemTree)
	}
	return nil
}

// EnrichBigMap -
func (set *Set) EnrichBigMap(bmd []*types.BigMapDiff) error {
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

// FindByName -
func (set *Set) FindByName(name string) Node {
	if set.GetName() == name {
		return set
	}
	return set.Type.FindByName(name)
}

// Docs -
func (set *Set) Docs(inferredName string) ([]Typedef, string, error) {
	name := getNameDocString(set, inferredName)
	typedef := Typedef{
		Name: name,
		Type: fmt.Sprintf("set(%s)", set.Type.GetPrim()),
		Args: make([]TypedefArg, 0),
	}
	if !isSimpleDocType(set.Type.GetPrim()) {
		docs, varName, err := set.Type.Docs(fmt.Sprintf("%s_item", name))
		if err != nil {
			return nil, "", err
		}

		typedef.Type = fmt.Sprintf("set(%s)", varName)
		result := []Typedef{typedef}
		for i := range docs {
			if !isFlatDocType(docs[i]) {
				result = append(result, docs[i])
			}
		}
		return result, typedef.Type, nil
	}
	return []Typedef{typedef}, typedef.Type, nil
}

// Distinguish -
func (set *Set) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Set)
	if !ok {
		return nil, nil
	}
	name := set.GetName()
	node := new(MiguelNode)
	node.Prim = set.Prim
	node.Type = set.Prim
	node.Name = &name

	d := getMatrix(set.Data, second.Data)
	children, err := mergeMatrix(d, len(set.Data), len(second.Data), set.Data, second.Data)
	if err != nil {
		return nil, err
	}

	node.Children = children

	return node, nil
}

// EqualType -
func (set *Set) EqualType(node Node) bool {
	if !set.Default.EqualType(node) {
		return false
	}
	second, ok := node.(*Set)
	if !ok {
		return false
	}
	return set.Type.EqualType(second.Type)
}

// FindPointers -
func (set *Set) FindPointers() map[int64]*BigMap {
	res := make(map[int64]*BigMap)
	for i := range set.Data {
		if b := set.Data[i].FindPointers(); b != nil {
			for k, v := range b {
				res[k] = v
			}
		}
	}
	return res
}

// Range -
func (set *Set) Range(handler func(node Node) error) error {
	if err := handler(set); err != nil {
		return err
	}
	return set.Type.Range(handler)
}

// GetJSONModel -
func (set *Set) GetJSONModel(model JSONModel) {
	if model == nil {
		return
	}
	arr := make([]JSONModel, len(set.Data))
	for i := range set.Data {
		set.Data[i].GetJSONModel(arr[i])
	}
	model[set.GetName()] = arr
}
