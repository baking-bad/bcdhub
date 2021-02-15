package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/pkg/errors"
)

const (
	leftKey  = 'L'
	rightKey = 'R'
)

func isOrPath(s string) bool {
	for _, chr := range s {
		if chr != leftKey && chr != rightKey {
			return false
		}
	}
	return true
}

// Or -
type Or struct {
	Default

	LeftType  Node
	RightType Node

	key byte
}

// NewOr -
func NewOr(depth int) *Or {
	return &Or{
		Default: NewDefault(consts.OR, -1, depth),
	}
}

// String -
func (or *Or) String() string {
	var s strings.Builder
	s.WriteString(or.Default.String())
	if or.key == leftKey {
		s.WriteString(strings.Repeat(consts.DefaultIndent, or.Depth))
		s.WriteString(or.LeftType.String())
	}
	if or.key == rightKey {
		s.WriteString(strings.Repeat(consts.DefaultIndent, or.Depth))
		s.WriteString(or.RightType.String())
	}
	return s.String()
}

// MarshalJSON -
func (or *Or) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.OR, or.Annots, or.LeftType, or.RightType)
}

// ParseType -
func (or *Or) ParseType(node *base.Node, id *int) error {
	if err := or.Default.ParseType(node, id); err != nil {
		return err
	}

	if len(node.Args) > 2 || len(node.Args) == 0 {
		return errors.Wrap(consts.ErrInvalidArgsCount, "Or.ParseValue")
	}

	for i := range node.Args {
		child, err := typingNode(node.Args[i], or.Depth, id)
		if err != nil {
			return err
		}
		if i == 0 {
			or.LeftType = child
		} else {
			or.RightType = child
		}
	}

	return nil
}

// ParseValue -
func (or *Or) ParseValue(node *base.Node) error {
	if len(node.Args) > 2 || len(node.Args) == 0 {
		return errors.Wrap(consts.ErrInvalidArgsCount, "Or.ParseValue")
	}

	switch node.Prim {
	case consts.Left:
		or.key = leftKey
		if err := or.LeftType.ParseValue(node.Args[0]); err != nil {
			return err
		}
	case consts.Right:
		or.key = rightKey
		if err := or.RightType.ParseValue(node.Args[0]); err != nil {
			return err
		}
	default:
		return errors.Wrap(consts.ErrInvalidPrim, "Or.ParseValue")
	}
	return nil
}

// ToMiguel -
func (or *Or) ToMiguel() (*MiguelNode, error) {
	node, err := or.Default.ToMiguel()
	if err != nil {
		return nil, err
	}

	node.Children = make([]*MiguelNode, 0)
	types := []Node{or.LeftType, or.RightType}
	for i := range types {
		child, err := types[i].ToMiguel()
		if err != nil {
			return nil, err
		}

		if or.Prim == types[i].GetPrim() {
			node.Children = append(node.Children, child.Children...)
		} else {
			node.Children = append(node.Children, child)
		}
	}

	node.Type = consts.TypeNamedEnum
	for i := range node.Children {
		if node.Children[i].Prim != consts.UNIT {
			node.Type = consts.TypeNamedUnion
			break
		}
	}

	return node, nil
}

// GetEntrypoints -
func (or *Or) GetEntrypoints() []string {
	e := make([]string, 0)
	for _, t := range []Node{or.LeftType, or.RightType} {
		e = append(e, t.GetEntrypoints()...)
	}
	return e
}

// ToBaseNode -
func (or *Or) ToBaseNode(optimized bool) (*base.Node, error) {
	node := new(base.Node)
	switch or.key {
	case leftKey:
		node.Prim = consts.Left
		arg, err := or.LeftType.ToBaseNode(optimized)
		if err != nil {
			return nil, err
		}
		node.Args = []*base.Node{arg}
	case rightKey:
		node.Prim = consts.Right
		arg, err := or.RightType.ToBaseNode(optimized)
		if err != nil {
			return nil, err
		}
		node.Args = []*base.Node{arg}
	default:
		return nil, errors.New("OR is not Left or Right")
	}
	return node, nil
}

// ToJSONSchema -
func (or *Or) ToJSONSchema() (*JSONSchema, error) {
	oneOf := make([]*JSONSchema, 0)
	for i, arg := range []Node{or.LeftType, or.RightType} {
		child, err := arg.ToJSONSchema()
		if err != nil {
			return nil, err
		}
		key := leftKey
		if i == 1 {
			key = rightKey
		}

		if child.Prim == consts.OR {
			for i := range child.OneOf {
				child.OneOf[i].Properties["schemaKey"].Const = string(key) + child.OneOf[i].Properties["schemaKey"].Const
				oneOf = append(oneOf, child.OneOf[i])
			}
		} else {

			item := &JSONSchema{
				Title: arg.GetName(),
				Properties: map[string]*JSONSchema{
					"schemaKey": {
						Type:  JSONSchemaTypeString,
						Const: string(key),
					},
				},
			}

			if child.Prim != consts.UNIT {
				for key, value := range child.Properties {
					item.Properties[key] = value
				}
			}

			oneOf = append(oneOf, item)
		}
	}

	return &JSONSchema{
		Type:  JSONSchemaTypeObject,
		Title: or.GetName(),
		Prim:  or.Prim,
		OneOf: oneOf,
	}, nil
}

// FromJSONSchema -
func (or *Or) FromJSONSchema(data map[string]interface{}) error {
	var orMap map[string]interface{}
	for key, val := range data {
		if key != or.GetName() {
			continue
		}
		arrVal, ok := val.(map[string]interface{})
		if !ok {
			return errors.Wrapf(consts.ErrInvalidType, "Or.FromJSONSchema %T", val)
		}
		orMap = arrVal
		break
	}

	schemaKey, ok := orMap["schemaKey"]
	if !ok {
		return errors.Wrap(consts.ErrJSONDataIsAbsent, "Or.FromJSONSchema")
	}

	sk, ok := schemaKey.(string)
	if !ok {
		return errors.Wrap(consts.ErrInvalidType, "Or.FromJSONSchema")
	}
	chr := sk[0]
	sk = sk[1:]
	orMap["schemaKey"] = sk
	switch chr {
	case leftKey:
		if or.LeftType.IsPrim(consts.OR) {
			orMap[or.LeftType.GetName()] = orMap
		}
		or.key = leftKey
		return or.LeftType.FromJSONSchema(orMap)
	case rightKey:
		if or.RightType.IsPrim(consts.OR) {
			orMap[or.RightType.GetName()] = orMap
		}
		or.key = rightKey
		return or.RightType.FromJSONSchema(orMap)
	default:
		return errors.Wrap(consts.ErrJSONDataIsAbsent, "Or.FromJSONSchema")
	}
}

// EnrichBigMap -
func (or *Or) EnrichBigMap(bmd []*base.BigMapDiff) error {
	switch or.key {
	case leftKey:
		return or.LeftType.EnrichBigMap(bmd)
	case rightKey:
		return or.RightType.EnrichBigMap(bmd)
	}

	return nil
}

// ToParameters -
func (or *Or) ToParameters() ([]byte, error) {
	var prim string
	var node Node
	switch or.key {
	case leftKey:
		prim = consts.Left
		node = or.LeftType
	case rightKey:
		prim = consts.Right
		node = or.RightType
	default:
		return nil, nil
	}

	var builder bytes.Buffer
	if _, err := builder.WriteString(fmt.Sprintf(`{"prim":"%s","args":[`, prim)); err != nil {
		return nil, err
	}
	b, err := node.ToParameters()
	if err != nil {
		return nil, err
	}
	if _, err := builder.Write(b); err != nil {
		return nil, err
	}
	if _, err := builder.WriteString(`]}`); err != nil {
		return nil, err
	}

	return builder.Bytes(), nil
}

// FindByName -
func (or *Or) FindByName(name string) Node {
	if or.GetName() == name {
		return or
	}

	if node := or.LeftType.FindByName(name); node != nil {
		return node
	}

	if node := or.RightType.FindByName(name); node != nil {
		return node
	}

	return nil
}

// Docs -
func (or *Or) Docs(inferredName string) ([]Typedef, string, error) {
	name := getNameDocString(or, inferredName)

	typedef := Typedef{
		Name: name,
		Type: or.Prim,
		Args: make([]TypedefArg, 0),
	}
	result := make([]Typedef, 0)

	types := []Node{or.LeftType, or.RightType}
	for i := range types {
		if isSimpleDocType(types[i].GetPrim()) {
			typedef.Args = append(typedef.Args, TypedefArg{
				Key:   types[i].GetName(),
				Value: types[i].GetPrim(),
			})
			continue
		}

		args, varName, err := types[i].Docs(name)
		if err != nil {
			return nil, "", err
		}
		if types[i].IsPrim(or.Prim) {
			typedef.Args = append(typedef.Args, args[0].Args...)
			for j := range args {
				if args[j].Type != or.Prim {
					result = append(result, args[j])
				}
			}
		} else {
			typedef.Args = append(typedef.Args, TypedefArg{
				Key:   types[i].GetName(),
				Value: varName,
			})
			for j := range args {
				if !isFlatDocType(args[j]) {
					result = append(result, args[j])
				}
			}
		}
	}
	result = append([]Typedef{typedef}, result...)
	return result, makeVarDocString(name), nil
}

// Compare -
func (or *Or) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*Or)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	var err error
	var res int
	switch {
	case or.key == leftKey && secondItem.key == leftKey:
		res, err = or.LeftType.Compare(secondItem.LeftType)
	case or.key == rightKey && secondItem.key == rightKey:
		res, err = or.RightType.Compare(secondItem.RightType)
	default:
		return 0, consts.ErrTypeIsNotComparable
	}
	if err != nil {
		return 0, err
	}
	return res, nil
}

// Distinguish -
func (or *Or) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Or)
	if !ok {
		return nil, nil
	}

	node, err := or.Default.ToMiguel()
	if err != nil {
		return nil, err
	}
	node.Children = make([]*MiguelNode, 0)

	switch {
	case or.key == leftKey && second.key == leftKey:
		child, err := or.LeftType.Distinguish(second.LeftType)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	case or.key == rightKey && second.key == rightKey:
		child, err := or.RightType.Distinguish(second.RightType)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	case or.key == 0 && second.key == 0:
	default:
		if second.key == rightKey {
			child, err := second.RightType.ToMiguel()
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, child)
			node.setDiffType(MiguelKindUpdate)
		}
		if second.key == leftKey {
			child, err := second.LeftType.ToMiguel()
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, child)
			node.setDiffType(MiguelKindUpdate)
		}
	}

	return node, nil
}

// EqualType -
func (or *Or) EqualType(node Node) bool {
	if !or.Default.EqualType(node) {
		return false
	}
	second, ok := node.(*Or)
	if !ok {
		return false
	}

	if !or.LeftType.EqualType(second.LeftType) {
		return false
	}

	return or.RightType.EqualType(second.RightType)
}
