package ast

import (
	"bytes"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/pkg/errors"
)

// Pair -
type Pair struct {
	Default
	Args []Node
}

// NewPair -
func NewPair(depth int) *Pair {
	return &Pair{
		Default: NewDefault(consts.PAIR, -1, depth),
	}
}

// String -
func (p *Pair) String() string {
	var s strings.Builder
	s.WriteString(p.Default.String())
	for i := range p.Args {
		s.WriteString(strings.Repeat(consts.DefaultIndent, p.Depth))
		s.WriteString(p.Args[i].String())
	}
	return s.String()
}

// MarshalJSON -
func (p *Pair) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.PAIR, p.Annots, p.Args...)
}

// ParseType -
func (p *Pair) ParseType(node *base.Node, id *int) error {
	if err := p.Default.ParseType(node, id); err != nil {
		return err
	}

	p.Args = make([]Node, 0)
	if len(node.Args) == 2 {
		for _, arg := range node.Args {
			child, err := typingNode(arg, p.Depth, id)
			if err != nil {
				return err
			}
			p.Args = append(p.Args, child)
		}
	} else if len(node.Args) > 2 {
		child, err := typingNode(node.Args[0], p.Depth, id)
		if err != nil {
			return err
		}
		p.Args = append(p.Args, child)

		newUntyped := &base.Node{
			Prim: consts.PAIR,
			Args: node.Args[1:],
		}
		pairChild, err := typingNode(newUntyped, p.Depth+1, id)
		if err != nil {
			return err
		}
		p.Args = append(p.Args, pairChild)
	}

	return nil
}

// ParseValue -
func (p *Pair) ParseValue(node *base.Node) error {
	switch {
	case len(node.Args) == 2:
		for i := range p.Args {
			if err := p.Args[i].ParseValue(node.Args[i]); err != nil {
				return err
			}
		}
		return nil
	case len(node.Args) > 2:
		if err := p.Args[0].ParseValue(node.Args[0]); err != nil {
			return err
		}

		newUntyped := &base.Node{
			Prim: consts.PAIR,
			Args: node.Args[1:],
		}
		return p.Args[1].ParseValue(newUntyped)
	default:
		return errors.Wrap(consts.ErrInvalidArgsCount, "Pair.ParseValue")
	}
}

// ToMiguel -
func (p *Pair) ToMiguel() (*MiguelNode, error) {
	name := p.GetTypeName()
	node := &MiguelNode{
		Prim:     p.Prim,
		Type:     consts.TypeNamedTuple,
		Name:     &name,
		Children: make([]*MiguelNode, 0),
	}

	for i := range p.Args {
		child, err := p.Args[i].ToMiguel()
		if err != nil {
			return nil, err
		}

		if p.Prim == p.Args[i].GetPrim() && strings.HasPrefix(*child.Name, "@") {
			node.Children = append(node.Children, child.Children...)
		} else {
			node.Children = append(node.Children, child)
		}
	}

	return node, nil
}

// ToBaseNode -
func (p *Pair) ToBaseNode(optimized bool) (*base.Node, error) {
	node := new(base.Node)
	node.Prim = consts.Pair
	node.Args = make([]*base.Node, 0)
	for i := range p.Args {
		arg, err := p.Args[i].ToBaseNode(optimized)
		if err != nil {
			return nil, err
		}
		node.Args = append(node.Args, arg)
	}
	return node, nil
}

// ToJSONSchema -
func (p *Pair) ToJSONSchema() (*JSONSchema, error) {
	s := &JSONSchema{
		Type:       JSONSchemaTypeObject,
		Properties: map[string]*JSONSchema{},
	}
	for _, arg := range p.Args {
		if err := setChildSchema(arg, false, s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// FromJSONSchema -
func (p *Pair) FromJSONSchema(data map[string]interface{}) error {
	for i := range p.Args {
		if err := p.Args[i].FromJSONSchema(data); err != nil {
			return err
		}
	}
	return nil
}

// EnrichBigMap -
func (p *Pair) EnrichBigMap(bmd []*types.BigMapDiff) error {
	for i := range p.Args {
		if err := p.Args[i].EnrichBigMap(bmd); err != nil {
			return err
		}
	}
	return nil
}

// ToParameters -
func (p *Pair) ToParameters() ([]byte, error) {
	var builder bytes.Buffer
	if _, err := builder.WriteString(`{"prim":"Pair","args":[`); err != nil {
		return nil, err
	}
	for i := range p.Args {
		if i > 0 {
			if err := builder.WriteByte(','); err != nil {
				return nil, err
			}
		}
		b, err := p.Args[i].ToParameters()
		if err != nil {
			return nil, err
		}
		if _, err := builder.Write(b); err != nil {
			return nil, err
		}
	}
	if _, err := builder.WriteString(`]}`); err != nil {
		return nil, err
	}
	return builder.Bytes(), nil
}

// FindByName -
func (p *Pair) FindByName(name string) Node {
	if p.GetName() == name {
		return p
	}
	for i := range p.Args {
		if node := p.Args[i].FindByName(name); node != nil {
			return node
		}
	}
	return nil
}

// Docs -
func (p *Pair) Docs(inferredName string) ([]Typedef, string, error) {
	typedef := Typedef{
		Name: getNameDocString(p, inferredName),
		Type: p.Prim,
		Args: make([]TypedefArg, 0),
	}
	result := make([]Typedef, 0)
	for i := range p.Args {
		args, varName, err := p.Args[i].Docs(typedef.Name)
		if err != nil {
			return nil, "", err
		}
		argName := p.Args[i].GetName()
		if isSimpleDocType(p.Args[i].GetPrim()) {
			typedef.Args = append(typedef.Args, TypedefArg{
				Key:   argName,
				Value: args[0].Type,
			})
			continue
		}

		if p.Args[i].IsPrim(p.Prim) {
			if p.Args[i].IsNamed() {
				typedef.Args = append(typedef.Args, TypedefArg{
					Key:   argName,
					Value: varName,
				})
				result = append(result, args...)
			} else {
				typedef.Args = append(typedef.Args, args[0].Args...)
				for j := range args {
					if isFlatDocType(args[j]) || isSimpleDocType(args[j].Type) {
						continue
					}
					if args[j].Type == p.Prim && args[j].Name == typedef.Name {
						continue
					}
					result = append(result, args[j])
				}
			}
		} else {
			typedef.Args = append(typedef.Args, TypedefArg{
				Key:   argName,
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

	return result, makeVarDocString(typedef.Name), nil
}

// Compare -
func (p *Pair) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*Pair)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	if len(secondItem.Args) != len(p.Args) {
		return 0, consts.ErrTypeIsNotComparable
	}

	for i := range p.Args {
		res, err := p.Args[i].Compare(secondItem.Args[i])
		if err != nil {
			return 0, err
		}
		if res != 0 {
			return res, nil
		}
	}

	return 0, nil
}

// Distinguish -
func (p *Pair) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Pair)
	if !ok {
		return nil, nil
	}

	node, err := p.Default.ToMiguel()
	if err != nil {
		return nil, err
	}
	node.Type = consts.TypeNamedTuple
	node.Children = make([]*MiguelNode, 0)

	for i := range p.Args {
		child, err := p.Args[i].Distinguish(second.Args[i])
		if err != nil {
			return nil, err
		}
		if (child.Prim != p.Prim) || (child.Prim == p.Prim && p.Args[i].IsNamed()) {
			node.Children = append(node.Children, child)
		} else {
			node.Children = append(node.Children, child.Children...)
		}
	}

	return node, nil
}

// EqualType -
func (p *Pair) EqualType(node Node) bool {
	if !p.Default.EqualType(node) {
		return false
	}
	second, ok := node.(*Pair)
	if !ok {
		return false
	}

	if len(p.Args) != len(second.Args) {
		return false
	}

	for i := range p.Args {
		if !p.Args[i].EqualType(second.Args[i]) {
			return false
		}
	}

	return true
}

// FindPointers -
func (p *Pair) FindPointers() map[int64]*BigMap {
	res := make(map[int64]*BigMap)
	for i := range p.Args {
		if b := p.Args[i].FindPointers(); b != nil {
			for k, v := range b {
				res[k] = v
			}
		}
	}
	return res
}

// Range -
func (p *Pair) Range(handler func(node Node) error) error {
	if err := handler(p); err != nil {
		return err
	}
	for i := range p.Args {
		if err := p.Args[i].Range(handler); err != nil {
			return err
		}
	}
	return nil
}

// GetJSONModel -
func (p *Pair) GetJSONModel(model JSONModel) {
	if model == nil {
		return
	}
	for i := range p.Args {
		p.Args[i].GetJSONModel(model)
	}
}
