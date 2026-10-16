package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/pkg/errors"
)

const (
	leftKey  = 'L'
	rightKey = 'R'
)

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

	if len(node.Args) != 2 {
		return errors.Wrapf(consts.ErrInvalidArgsCount, "Or.ParseType: expects 2 args, got %d", len(node.Args))
	}

	left, err := typeNode(node.Args[0], or.Depth, id)
	if err != nil {
		return err
	}
	right, err := typeNode(node.Args[1], or.Depth, id)
	if err != nil {
		return err
	}
	or.LeftType = left
	or.RightType = right

	return nil
}

// ParseValue -
func (or *Or) ParseValue(node *base.Node) error {
	var (
		key   byte
		child Node
	)
	switch node.Prim {
	case consts.Left:
		key, child = leftKey, or.LeftType
	case consts.Right:
		key, child = rightKey, or.RightType
	default:
		return errors.Wrapf(consts.ErrInvalidPrim, "Or.ParseValue: expected Left or Right, got '%s'", node.Prim)
	}

	if len(node.Args) != 1 {
		return errors.Wrapf(consts.ErrInvalidArgsCount, "Or.ParseValue: %s expects 1 arg, got %d", node.Prim, len(node.Args))
	}
	if err := child.ParseValue(node.Args[0]); err != nil {
		return err
	}
	or.key = key
	return nil
}

// ToMiguel -
func (or *Or) ToMiguel() (*MiguelNode, error) {
	node, err := or.Default.ToMiguel()
	if err != nil {
		return nil, err
	}

	node.Children = make([]*MiguelNode, 0)
	typ := or.LeftType
	if or.key == rightKey {
		typ = or.RightType
	}
	child, err := typ.ToMiguel()
	if err != nil {
		return nil, err
	}

	if or.Prim == typ.GetPrim() {
		node.Children = append(node.Children, child.Children...)
	} else {
		node.Children = append(node.Children, child)
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

		if arg.IsPrim(consts.OR) {
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

			switch {
			case arg.IsPrim(consts.OPTION) || arg.IsPrim(consts.PAIR):
				item.Properties[arg.GetName()] = child
			case arg.IsPrim(consts.UNIT):
			default:
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
//
// `schemaKey` is a path of 'L'/'R' chars: one char per nested `or`. Input `data` is not modified.
func (or *Or) FromJSONSchema(data map[string]interface{}) error {
	value, ok := data[or.GetName()]
	if !ok {
		return errors.Wrap(consts.ErrJSONDataIsAbsent, "Or.FromJSONSchema")
	}
	orMap, ok := value.(map[string]interface{})
	if !ok {
		return errors.Wrapf(consts.ErrValidation, "Or.FromJSONSchema: expected object in '%s', got %T", or.GetName(), value)
	}

	schemaKey, ok := orMap["schemaKey"]
	if !ok {
		return errors.Wrap(consts.ErrJSONDataIsAbsent, "Or.FromJSONSchema: schemaKey")
	}
	sk, ok := schemaKey.(string)
	if !ok {
		return errors.Wrapf(consts.ErrValidation, "Or.FromJSONSchema: schemaKey should be a string, got %T", schemaKey)
	}
	if sk == "" {
		return errors.Wrapf(consts.ErrValidation, "Or.FromJSONSchema: empty schemaKey in '%s'", or.GetName())
	}

	var child Node
	switch sk[0] {
	case leftKey:
		child = or.LeftType
	case rightKey:
		child = or.RightType
	default:
		return errors.Wrapf(consts.ErrValidation, "Or.FromJSONSchema: invalid schemaKey '%s'", sk)
	}

	rest := sk[1:]
	childData := make(map[string]interface{}, len(orMap))
	for key, val := range orMap {
		if key != "schemaKey" {
			childData[key] = val
		}
	}

	if child.IsPrim(consts.OR) {
		if rest == "" {
			return errors.Wrapf(consts.ErrValidation, "Or.FromJSONSchema: schemaKey '%s' is too short", sk)
		}
		childData["schemaKey"] = rest
		childData = map[string]interface{}{
			child.GetName(): childData,
		}
	} else if rest != "" {
		return errors.Wrapf(consts.ErrValidation, "Or.FromJSONSchema: schemaKey '%s' is too long", sk)
	}

	or.key = sk[0]
	return child.FromJSONSchema(childData)
}

// EnrichBigMap -
func (or *Or) EnrichBigMap(bmd []*types.BigMapDiff) error {
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
	if _, err := fmt.Fprintf(&builder, `{"prim":"%s","args":[`, prim); err != nil {
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
func (or *Or) FindByName(name string, isEntrypoint bool) Node {
	if or.GetName() == name {
		return or
	}

	if node := or.LeftType.FindByName(name, isEntrypoint); node != nil {
		return node
	}

	if node := or.RightType.FindByName(name, isEntrypoint); node != nil {
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
	case or.key == leftKey && secondItem.key == rightKey:
		res = -1
	case or.key == rightKey && secondItem.key == leftKey:
		res = 1
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
		child, err := distinguishNodes(or.LeftType, second.LeftType)
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	case or.key == rightKey && second.key == rightKey:
		child, err := distinguishNodes(or.RightType, second.RightType)
		if err != nil {
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	case or.key == 0 && second.key == 0:
	default:
		if second.key == rightKey {
			child, err := or.LeftType.ToMiguel()
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, child)
			node.setDiffType(MiguelKindUpdate)
		}
		if second.key == leftKey {
			child, err := or.RightType.ToMiguel()
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

// FindPointers -
func (or *Or) FindPointers() map[int64]*BigMap {
	switch or.key {
	case leftKey:
		return or.LeftType.FindPointers()
	case rightKey:
		return or.RightType.FindPointers()
	}
	return nil
}

// GetJSONModel -
func (or *Or) GetJSONModel(model JSONModel) {
	if model == nil {
		return
	}

	var node Node
	switch or.key {
	case leftKey:
		node = or.LeftType
	case rightKey:
		node = or.RightType
	default:
		return
	}

	item := make(JSONModel)
	node.GetJSONModel(item)

	if node.IsPrim(or.Prim) {
		val, ok := model[node.GetName()]
		if !ok {
			return
		}
		child := val.(JSONModel)
		child["schemaKey"] = fmt.Sprintf("%s%v", string(or.key), child["schemaKey"])
		delete(model, node.GetName())
		model[or.GetName()] = child
	} else {
		newModel := JSONModel{
			"schemaKey": string(or.key),
		}
		for key := range item {
			newModel[key] = item[key]
		}
		model[or.GetName()] = newModel
	}
}
