package ast

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/pkg/errors"
)

// Map -
type Map struct {
	Default
	KeyType   Node
	ValueType Node

	Data *OrderedMap
}

// NewMap -
func NewMap(depth int) *Map {
	return &Map{
		Default: NewDefault(consts.MAP, 2, depth),
		Data:    NewOrderedMap(),
	}
}

// String -
func (m *Map) String() string {
	var s strings.Builder

	s.WriteString(m.Default.String())
	if m.Data.Len() > 0 {
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
	} else {
		s.WriteString(strings.Repeat(consts.DefaultIndent, m.Depth))
		s.WriteString(m.KeyType.String())
		s.WriteString(strings.Repeat(consts.DefaultIndent, m.Depth))
		s.WriteString(m.ValueType.String())
	}

	return s.String()
}

// MarshalJSON -
func (m *Map) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.MAP, m.Annots, m.KeyType, m.ValueType)
}

// ParseType -
func (m *Map) ParseType(node *base.Node, id *int) error {
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
func (m *Map) ParseValue(node *base.Node) error {
	if node.Prim != consts.PrimArray {
		return errors.Wrap(consts.ErrInvalidPrim, "Map.ParseValue")
	}

	return createMapFromElts(node.Args, m.KeyType, m.ValueType, m.Data)
}

// ToMiguel -
func (m *Map) ToMiguel() (*MiguelNode, error) {
	node, err := m.Default.ToMiguel()
	if err != nil {
		return nil, err
	}

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

// ToBaseNode -
func (m *Map) ToBaseNode(optimized bool) (*base.Node, error) {
	return mapToBaseNodes(m.Data, optimized)
}

// ToJSONSchema -
func (m *Map) ToJSONSchema() (*JSONSchema, error) {
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
func (m *Map) FromJSONSchema(data map[string]interface{}) error {
	m.Data = NewOrderedMap()
	var arr []interface{}
	for key := range data {
		if key == m.GetName() {
			val := data[key]
			arrVal, ok := val.([]interface{})
			if !ok {
				return errors.Wrapf(consts.ErrInvalidType, "Map.FromJSONSchema %T", val)
			}
			arr = arrVal
			break
		}
	}
	if arr == nil {
		return errors.Wrap(consts.ErrJSONDataIsAbsent, "Map.FromJSONSchema")
	}

	for i := range arr {
		item, ok := arr[i].(map[string]interface{})
		if !ok {
			return errors.Wrap(consts.ErrValidation, "Map.FromJSONSchema")
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
func (m *Map) EnrichBigMap(bmd []*types.BigMapDiff) error {
	return m.Data.Range(func(key, value Comparable) (bool, error) {
		if err := key.(Node).EnrichBigMap(bmd); err != nil {
			return true, err
		}
		if err := value.(Node).EnrichBigMap(bmd); err != nil {
			return true, err
		}
		return false, nil
	})
}

// ToParameters -
func (m *Map) ToParameters() ([]byte, error) {
	return buildMapParameters(m.Data)
}

// FindByName -
func (m *Map) FindByName(name string) Node {
	if m.GetName() == name {
		return m
	}
	node := m.KeyType.FindByName(name)
	if node != nil {
		return node
	}
	return m.ValueType.FindByName(name)
}

// Docs -
func (m *Map) Docs(inferredName string) ([]Typedef, string, error) {
	name := getNameDocString(m, inferredName)
	typedef := Typedef{
		Name: name,
		Type: fmt.Sprintf("map(%s, %s)", m.KeyType.GetPrim(), m.ValueType.GetPrim()),
		Args: make([]TypedefArg, 0),
	}

	if isSimpleDocType(m.KeyType.GetPrim()) && isSimpleDocType(m.ValueType.GetPrim()) {
		return []Typedef{typedef}, typedef.Type, nil
	}
	keyDocs, keyVarName, err := m.KeyType.Docs(fmt.Sprintf("%s_key", name))
	if err != nil {
		return nil, "", err
	}
	typedef.Args = append(typedef.Args, TypedefArg{Key: keyDocs[0].Name, Value: keyVarName})

	valDocs, valVarName, err := m.ValueType.Docs(fmt.Sprintf("%s_value", name))
	if err != nil {
		return nil, "", err
	}
	typedef.Args = append(typedef.Args, TypedefArg{Key: valDocs[0].Name, Value: valVarName})

	typedef.Type = fmt.Sprintf("map(%s, %s)", keyVarName, valVarName)
	result := []Typedef{typedef}
	result = append(result, keyDocs...)
	result = append(result, valDocs...)

	return result, typedef.Type, nil
}

// Distinguish -
func (m *Map) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Map)
	if !ok {
		return nil, nil
	}
	name := m.GetName()
	node := new(MiguelNode)
	node.Prim = m.Prim
	node.Type = m.Prim
	node.Name = &name
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
func (m *Map) EqualType(node Node) bool {
	if !m.Default.EqualType(node) {
		return false
	}
	second, ok := node.(*Map)
	if !ok {
		return false
	}

	if !m.KeyType.EqualType(second.KeyType) {
		return false
	}

	return m.ValueType.EqualType(second.ValueType)
}

func createMapFromElts(args []*base.Node, keyType, valueType Node, data *OrderedMap) error {
	if data == nil {
		data = NewOrderedMap()
	}

	for i := range args {
		elt := args[i]
		if elt.Prim != consts.Elt {
			return errors.Wrap(consts.ErrInvalidPrim, "createMapFromElts")
		}
		if len(elt.Args) != 2 {
			return errors.Wrap(consts.ErrInvalidArgsCount, "createMapFromElts")
		}

		if elt.Args[1] != nil {
			key := Copy(keyType)
			if err := key.ParseValue(elt.Args[0]); err != nil {
				return err
			}
			val := Copy(valueType)
			if err := val.ParseValue(elt.Args[1]); err != nil {
				return err
			}
			if err := data.Add(key, val); err != nil {
				return err
			}
		}
	}
	return nil
}

func getMapKeyName(node *MiguelNode) (s string, err error) {
	switch kv := node.Value.(type) {
	case string:
		if kv == "" {
			kv = `""`
		}
		s = kv
	case int, int64:
		s = fmt.Sprintf("%d", kv)
	case bool:
		s = fmt.Sprintf("%t", kv)
	case map[string]interface{}:
		s = fmt.Sprintf("%v", kv["miguel_value"])
	case []interface{}:
		s = ""
		for i, item := range kv {
			val := item.(map[string]interface{})
			if i != 0 {
				s += "@"
			}
			s += fmt.Sprintf("%v", val["miguel_value"])
		}
	default:
		err = errors.Errorf("Invalid map key type: %v %T", node, node)
	}
	return
}
