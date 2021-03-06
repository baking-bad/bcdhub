package ast

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/pkg/errors"
)

// BigMap -
type BigMap struct {
	Default

	KeyType   Node
	ValueType Node

	Data *OrderedMap
	Ptr  *int64

	diffs []*types.BigMapDiff
}

// NewBigMap -
func NewBigMap(depth int) *BigMap {
	return &BigMap{
		Default: NewDefault(consts.BIGMAP, 2, depth),
		Data:    NewOrderedMap(),
		diffs:   make([]*types.BigMapDiff, 0),
	}
}

// String -
func (m *BigMap) String() string {
	var s strings.Builder

	s.WriteString(m.Default.String())
	switch {
	case m.Ptr != nil:
		s.WriteString(strings.Repeat(consts.DefaultIndent, m.Depth))
		s.WriteString(fmt.Sprintf("Ptr=%d\n", *m.Ptr))
	case m.Data.Len() > 0:
		_ = m.Data.Range(func(key, val Comparable) (bool, error) {
			s.WriteString(strings.Repeat(consts.DefaultIndent, m.Depth))
			s.WriteByte('{')
			s.WriteByte('\n')
			s.WriteString(strings.Repeat(consts.DefaultIndent, m.Depth+1))
			s.WriteString(key.(Node).String())
			s.WriteString(strings.Repeat(consts.DefaultIndent, m.Depth+1))
			s.WriteString(val.(Node).String())
			s.WriteString(strings.Repeat(consts.DefaultIndent, m.Depth))
			s.WriteByte('}')
			s.WriteByte('\n')
			return false, nil
		})
	default:
		s.WriteString(strings.Repeat(consts.DefaultIndent, m.Depth))
		s.WriteString(m.KeyType.String())
		s.WriteString(strings.Repeat(consts.DefaultIndent, m.Depth))
		s.WriteString(m.ValueType.String())
	}

	return s.String()
}

// MarshalJSON -
func (m *BigMap) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.BIGMAP, m.Annots, m.KeyType, m.ValueType)
}

// ParseType -
func (m *BigMap) ParseType(node *base.Node, id *int) error {
	if err := m.Default.ParseType(node, id); err != nil {
		return err
	}

	keyType, err := typingNode(node.Args[0], m.Depth, id)
	if err != nil {
		return err
	}
	m.KeyType = keyType

	valType, err := typingNode(node.Args[1], m.Depth, id)
	if err != nil {
		return err
	}
	m.ValueType = valType

	return nil
}

// ParseValue -
func (m *BigMap) ParseValue(node *base.Node) error {
	switch {
	case node.Prim == consts.PrimArray:
		return createMapFromElts(node.Args, m.KeyType, m.ValueType, m.Data)
	case node.IntValue != nil:
		ptr := node.IntValue.Int64()
		m.Ptr = &ptr
	default:
		return errors.Wrap(consts.ErrInvalidPrim, fmt.Sprintf("BigMap.ParseValue (%s)", node.Prim))
	}
	return nil
}

// ToMiguel -
func (m *BigMap) ToMiguel() (*MiguelNode, error) {
	node, err := m.Default.ToMiguel()
	if err != nil {
		return nil, err
	}

	switch {
	case m.Ptr != nil:
		node.Value = *m.Ptr
		return node, nil
	default:
		node.Children = make([]*MiguelNode, 0)
		handler := func(key, value Comparable) (bool, error) {
			keyChild, err := key.(Node).ToMiguel()
			if err != nil {
				return true, err
			}
			if value != nil {
				child, err := value.(Node).ToMiguel()
				if err != nil {
					return true, err
				}

				name, err := getMapKeyName(keyChild)
				if err != nil {
					return true, err
				}
				child.Name = &name
				node.Children = append(node.Children, child)
			}
			return false, nil
		}

		err = m.Data.Range(handler)
		return node, err
	}

}

// ToBaseNode -
func (m *BigMap) ToBaseNode(optimized bool) (*base.Node, error) {
	if m.Data.Len() > 0 {
		return mapToBaseNodes(m.Data, optimized)
	}
	if m.Ptr != nil {
		return toBaseNodeInt(types.NewBigInt(*m.Ptr)), nil
	}
	return nil, nil
}

// ToJSONSchema -
func (m *BigMap) ToJSONSchema() (*JSONSchema, error) {
	s := &JSONSchema{
		Type:    JSONSchemaTypeArray,
		Title:   m.GetName(),
		Default: make([]interface{}, 0),
		Items: &SchemaKey{
			Type:       JSONSchemaTypeObject,
			Required:   make([]string, 0),
			Properties: make(map[string]*JSONSchema),
		},
	}

	if err := setChildSchemaForMap(m.KeyType, true, s); err != nil {
		return nil, err
	}

	if err := setChildSchemaForMap(m.ValueType, false, s); err != nil {
		return nil, err
	}

	return wrapObject(s), nil
}

// FromJSONSchema -
func (m *BigMap) FromJSONSchema(data map[string]interface{}) error {
	m.Data = NewOrderedMap()
	var arr []interface{}
	for key := range data {
		if key == m.GetName() {
			val := data[key]
			arrVal, ok := val.([]interface{})
			if !ok {
				return errors.Wrapf(consts.ErrInvalidType, "BigMap.FromJSONSchema %T", val)
			}
			arr = arrVal
			break
		}
	}
	if arr == nil {
		return errors.Wrap(consts.ErrJSONDataIsAbsent, "BigMap.FromJSONSchema")
	}

	for i := range arr {
		item, ok := arr[i].(map[string]interface{})
		if !ok {
			return errors.Wrap(consts.ErrValidation, "BigMap.FromJSONSchema")
		}
		keyTree := Copy(m.KeyType)
		if err := keyTree.FromJSONSchema(item); err != nil {
			return err
		}
		valTree := Copy(m.ValueType)
		if err := valTree.FromJSONSchema(item); err != nil {
			return err
		}
		if err := m.Data.Add(keyTree, valTree); err != nil {
			return err
		}
	}
	return nil
}

// EnrichBigMap -
func (m *BigMap) EnrichBigMap(bmd []*types.BigMapDiff) error {
	for i := range bmd {
		if m.Ptr != nil && bmd[i].Ptr == *m.Ptr {
			key, err := m.makeNodeFromBytes(m.KeyType, bmd[i].Key)
			if err != nil {
				return err
			}
			var val Node
			if bmd[i].Value != nil {
				val, err = m.makeNodeFromBytes(m.ValueType, bmd[i].Value)
				if err != nil {
					return err
				}
			}
			if err := m.Data.Add(key, val); err != nil {
				return err
			}
			m.diffs = append(m.diffs, bmd[i])
		}
	}
	return nil
}

// ToParameters -
func (m *BigMap) ToParameters() ([]byte, error) {
	return buildMapParameters(m.Data)
}

// FindByName -
func (m *BigMap) FindByName(name string, isEntrypoint bool) Node {
	if m.GetName() == name {
		return m
	}
	if isEntrypoint {
		return nil
	}
	node := m.KeyType.FindByName(name, isEntrypoint)
	if node != nil {
		return node
	}
	return m.ValueType.FindByName(name, isEntrypoint)
}

func (m *BigMap) makeNodeFromBytes(typ Node, data []byte) (Node, error) {
	value := Copy(typ)
	var node base.Node
	if err := json.Unmarshal(data, &node); err != nil {
		return nil, err
	}
	if err := value.ParseValue(&node); err != nil {
		return nil, err
	}
	return value, nil
}

// Docs -
func (m *BigMap) Docs(inferredName string) ([]Typedef, string, error) {
	typedef := Typedef{
		Name: m.GetName(),
		Type: fmt.Sprintf("big_map(%s, %s)", m.KeyType.GetPrim(), m.ValueType.GetPrim()),
		Args: make([]TypedefArg, 0),
	}

	if isSimpleDocType(m.KeyType.GetPrim()) && isSimpleDocType(m.ValueType.GetPrim()) {
		return []Typedef{typedef}, typedef.Type, nil
	}
	keyDocs, keyVarName, err := m.KeyType.Docs(fmt.Sprintf("%s_key", typedef.Name))
	if err != nil {
		return nil, "", err
	}

	valDocs, valVarName, err := m.ValueType.Docs(fmt.Sprintf("%s_value", typedef.Name))
	if err != nil {
		return nil, "", err
	}

	typedef.Type = fmt.Sprintf("big_map(%s, %s)", keyVarName, valVarName)
	result := []Typedef{typedef}
	if strings.HasPrefix(keyVarName, "$") {
		typedef.Args = append(typedef.Args, TypedefArg{Key: keyDocs[0].Name, Value: keyVarName})
		result = append(result, keyDocs...)
	}
	if strings.HasPrefix(valVarName, "$") {
		typedef.Args = append(typedef.Args, TypedefArg{Key: valDocs[0].Name, Value: valVarName})
		result = append(result, valDocs...)
	}

	return result, typedef.Type, nil
}

// Distinguish -
func (m *BigMap) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*BigMap)
	if !ok {
		return nil, nil
	}
	name := m.GetName()
	node := new(MiguelNode)
	node.Prim = m.Prim
	node.Type = m.Prim
	node.Name = &name
	if m.Ptr != nil {
		node.Value = m.Ptr
	}
	node.Children = make([]*MiguelNode, 0)

	err := m.Data.Range(func(key, value Comparable) (bool, error) {
		keyChild, err := key.(Node).ToMiguel()
		if err != nil {
			return true, err
		}
		name, err := getMapKeyName(keyChild)
		if err != nil {
			return true, err
		}

		val, ok := second.Data.Get(key)
		if !ok {
			child, err := value.(Node).ToMiguel()
			if err != nil {
				return true, err
			}
			child.setDiffType(MiguelKindCreate)
			child.Name = &name
			node.Children = append(node.Children, child)
			return false, nil
		}

		child, err := value.(Node).Distinguish(val.(Node))
		if err != nil {
			return true, err
		}
		child.Name = &name
		node.Children = append(node.Children, child)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	err = second.Data.Range(func(key, value Comparable) (bool, error) {
		if _, ok := m.Data.Get(key); !ok {
			child, err := value.(Node).ToMiguel()
			if err != nil {
				return true, err
			}
			keyChild, err := key.(Node).ToMiguel()
			if err != nil {
				return true, err
			}
			name, err := getMapKeyName(keyChild)
			if err != nil {
				return true, err
			}
			child.setDiffType(MiguelKindDelete)
			child.Name = &name
			node.Children = append(node.Children, child)
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return node, nil
}

// EqualType -
func (m *BigMap) EqualType(node Node) bool {
	if !m.Default.EqualType(node) {
		return false
	}
	second, ok := node.(*BigMap)
	if !ok {
		return false
	}

	if !m.KeyType.EqualType(second.KeyType) {
		return false
	}

	return m.ValueType.EqualType(second.ValueType)
}

// SetKeyType -
func (m *BigMap) SetKeyType(data []byte) error {
	var node UntypedAST
	if err := json.Unmarshal(data, &node); err != nil {
		return err
	}
	typ, err := node.ToTypedAST()
	if err != nil {
		return err
	}
	m.KeyType = typ.Nodes[0]
	return nil
}

// SetValueType -
func (m *BigMap) SetValueType(data []byte) error {
	var node UntypedAST
	if err := json.Unmarshal(data, &node); err != nil {
		return err
	}
	typ, err := node.ToTypedAST()
	if err != nil {
		return err
	}
	m.ValueType = typ.Nodes[0]
	return nil
}

// AddDiffs -
func (m *BigMap) AddDiffs(diffs ...*types.BigMapDiff) {
	m.diffs = append(m.diffs, diffs...)
}

// GetDiffs -
func (m *BigMap) GetDiffs() []*types.BigMapDiff {
	return m.diffs
}

// FindPointers -
func (m *BigMap) FindPointers() map[int64]*BigMap {
	res := make(map[int64]*BigMap)
	if m.Ptr != nil {
		res[*m.Ptr] = m
	}

	if err := m.Data.Range(func(_, value Comparable) (bool, error) {
		if value == nil {
			return false, nil
		}
		node := value.(Node)
		if m := node.FindPointers(); m != nil {
			for k, v := range m {
				res[k] = v
			}
		}
		return false, nil
	}); err != nil {
		return nil
	}
	return res
}

// Range -
func (m *BigMap) Range(handler func(node Node) error) error {
	if err := handler(m); err != nil {
		return err
	}
	if err := m.KeyType.Range(handler); err != nil {
		return err
	}
	return m.ValueType.Range(handler)
}

// GetJSONModel -
func (m *BigMap) GetJSONModel(model JSONModel) {
	if model == nil {
		return
	}
	arr := make([]JSONModel, 0)
	_ = m.Data.Range(func(key, value Comparable) (bool, error) {
		item := make(JSONModel)
		key.(Node).GetJSONModel(item)
		value.(Node).GetJSONModel(item)
		arr = append(arr, item)
		return false, nil
	})
	model[m.GetName()] = arr
}
