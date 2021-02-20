package ast

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/pkg/errors"
)

// List -
type List struct {
	Default

	Type Node

	Data []Node
}

// NewList -
func NewList(depth int) *List {
	return &List{
		Default: NewDefault(consts.LIST, -1, depth),
	}
}

// MarshalJSON -
func (list *List) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.LIST, list.Annots, list.Type)
}

// String -
func (list *List) String() string {
	var s strings.Builder

	s.WriteString(list.Default.String())
	if len(list.Data) > 0 {
		for i := range list.Data {
			s.WriteString(strings.Repeat(consts.DefaultIndent, list.Depth))
			s.WriteByte('{')
			s.WriteByte('\n')
			s.WriteString(strings.Repeat(consts.DefaultIndent, list.Depth+1))
			s.WriteString(list.Data[i].String())
			s.WriteString(strings.Repeat(consts.DefaultIndent, list.Depth))
			s.WriteByte('}')
			s.WriteByte('\n')
		}
	} else {
		s.WriteString(strings.Repeat(consts.DefaultIndent, list.Depth))
		s.WriteString(list.Type.String())
	}
	return s.String()
}

// ParseType -
func (list *List) ParseType(node *base.Node, id *int) error {
	if err := list.Default.ParseType(node, id); err != nil {
		return err
	}

	typ, err := typingNode(node.Args[0], list.Depth, id)
	if err != nil {
		return err
	}
	list.Type = typ

	return nil
}

// ParseValue -
func (list *List) ParseValue(node *base.Node) error {
	if node.Prim != consts.PrimArray {
		return errors.Wrap(consts.ErrInvalidPrim, "List.ParseValue")
	}

	list.Data = make([]Node, 0)

	for i := range node.Args {
		item := Copy(list.Type)
		if err := item.ParseValue(node.Args[i]); err != nil {
			return err
		}
		list.Data = append(list.Data, item)
	}

	return nil
}

// ToMiguel -
func (list *List) ToMiguel() (*MiguelNode, error) {
	node, err := list.Default.ToMiguel()
	if err != nil {
		return nil, err
	}

	node.Children = make([]*MiguelNode, 0)
	for i := range list.Data {
		child, err := list.Data[i].ToMiguel()
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	}

	return node, nil
}

// ToBaseNode -
func (list *List) ToBaseNode(optimized bool) (*base.Node, error) {
	return arrayToBaseNode(list.Data, optimized)
}

// ToJSONSchema -
func (list *List) ToJSONSchema() (*JSONSchema, error) {
	s := &JSONSchema{
		Prim:    list.Prim,
		Type:    JSONSchemaTypeArray,
		Title:   list.GetName(),
		Default: make([]interface{}, 0),
		Items: &SchemaKey{
			Type:       JSONSchemaTypeObject,
			Required:   make([]string, 0),
			Properties: make(map[string]*JSONSchema),
		},
	}

	if err := setChildSchema(list.Type, true, s); err != nil {
		return nil, err
	}

	return wrapObject(s), nil
}

// FromJSONSchema -
func (list *List) FromJSONSchema(data map[string]interface{}) error {
	var arr []interface{}
	for key := range data {
		if key == list.GetName() {
			val := data[key]
			arrVal, ok := val.([]interface{})
			if !ok {
				return errors.Wrapf(consts.ErrInvalidType, "List.FromJSONSchema %T", val)
			}
			arr = arrVal
			break
		}
	}
	if arr == nil {
		return errors.Wrap(consts.ErrJSONDataIsAbsent, "List.FromJSONSchema")
	}

	for i := range arr {
		item, ok := arr[i].(map[string]interface{})
		if !ok {
			return errors.Wrap(consts.ErrValidation, "List.FromJSONSchema")
		}
		itemTree := Copy(list.Type)
		if err := itemTree.FromJSONSchema(item); err != nil {
			return err
		}

		list.Data = append(list.Data, itemTree)
	}
	return nil
}

// EnrichBigMap -
func (list *List) EnrichBigMap(bmd []*types.BigMapDiff) error {
	for i := range list.Data {
		if err := list.Data[i].EnrichBigMap(bmd); err != nil {
			return err
		}
	}
	return nil
}

// ToParameters -
func (list *List) ToParameters() ([]byte, error) {
	return buildListParameters(list.Data)
}

// FindByName -
func (list *List) FindByName(name string) Node {
	if list.GetName() == name {
		return list
	}
	return list.Type.FindByName(name)
}

// Docs -
func (list *List) Docs(inferredName string) ([]Typedef, string, error) {
	name := getNameDocString(list, inferredName)
	typedef := Typedef{
		Name: name,
		Type: fmt.Sprintf("list(%s)", list.Type.GetPrim()),
		Args: make([]TypedefArg, 0),
	}
	if !isSimpleDocType(list.Type.GetPrim()) {
		docs, varName, err := list.Type.Docs(fmt.Sprintf("%s_item", name))
		if err != nil {
			return nil, "", err
		}

		typedef.Type = fmt.Sprintf("list(%s)", varName)
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
func (list *List) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*List)
	if !ok {
		return nil, nil
	}
	name := list.GetName()
	node := new(MiguelNode)
	node.Prim = list.Prim
	node.Type = list.Prim
	node.Name = &name

	d := getMatrix(list.Data, second.Data)
	children, err := mergeMatrix(d, len(list.Data), len(second.Data), list.Data, second.Data)
	if err != nil {
		return nil, err
	}

	node.Children = children

	return node, nil
}

// EqualType -
func (list *List) EqualType(node Node) bool {
	if !list.Default.EqualType(node) {
		return false
	}
	second, ok := node.(*List)
	if !ok {
		return false
	}
	return list.Type.EqualType(second.Type)
}

// FindPointers -
func (list *List) FindPointers() map[int64]*BigMap {
	res := make(map[int64]*BigMap)
	for i := range list.Data {
		if b := list.Data[i].FindPointers(); b != nil {
			for k, v := range b {
				res[k] = v
			}
		}
	}
	return res
}

// Range -
func (list *List) Range(handler func(node Node) error) error {
	if err := handler(list); err != nil {
		return err
	}
	return list.Type.Range(handler)
}

func mergeMatrix(d [][]int, i, j int, first, second []Node) ([]*MiguelNode, error) {
	children := make([]*MiguelNode, 0)
	var err error
	if i == 0 && j == 0 {
		return children, nil
	}
	if i == 0 {
		for idx := 0; idx < j; idx++ {
			item, err := second[idx].ToMiguel()
			if err != nil {
				return nil, err
			}
			item.setDiffType(MiguelKindDelete)
			children = append(children, item)
		}
		return children, nil
	}
	if j == 0 {
		for idx := 0; idx < i; idx++ {
			item, err := first[idx].ToMiguel()
			if err != nil {
				return nil, err
			}
			item.setDiffType(MiguelKindCreate)
			children = append(children, item)
		}
		return children, nil
	}
	left := d[i][j-1]
	up := d[i-1][j]
	upleft := d[i-1][j-1]

	if upleft <= up && upleft <= left {
		if upleft == d[i][j] {
			children, err = mergeMatrix(d, i-1, j-1, first, second)
			if err != nil {
				return nil, err
			}
			item, err := first[i-1].Distinguish(second[j-1])
			if err != nil {
				return nil, err
			}
			children = append(children, item)
		} else {
			children, err = mergeMatrix(d, i-1, j-1, first, second)
			if err != nil {
				return nil, err
			}
			item, err := first[i-1].Distinguish(second[j-1])
			if err != nil {
				return nil, err
			}
			item.setDiffType(MiguelKindUpdate)
			item.From = fmt.Sprintf("%v", second[j-1].GetValue())
			children = append(children, item)
		}
	} else {
		if left <= upleft && left <= up {
			children, err = mergeMatrix(d, i, j-1, first, second)
			if err != nil {
				return nil, err
			}
			item, err := second[j-1].ToMiguel()
			if err != nil {
				return nil, err
			}
			item.setDiffType(MiguelKindDelete)
			children = append(children, item)
		} else {
			children, err = mergeMatrix(d, i-1, j, first, second)
			if err != nil {
				return nil, err
			}
			item, err := first[i-1].ToMiguel()
			if err != nil {
				return nil, err
			}
			item.setDiffType(MiguelKindCreate)
			children = append(children, item)
		}
	}
	return children, nil
}

func getMatrix(first, second []Node) [][]int {
	n := len(second)
	m := len(first)

	d := make([][]int, m+1)
	for i := 0; i < m+1; i++ {
		d[i] = make([]int, n+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}

	for j := 1; j < n+1; j++ {
		for i := 1; i < m+1; i++ {
			cost := 0
			if !first[i-1].Equal(second[j-1]) {
				cost = 1
			}

			d[i][j] = min(min(d[i-1][j]+1, d[i][j-1]+1), d[i-1][j-1]+cost)
		}
	}
	return d
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
