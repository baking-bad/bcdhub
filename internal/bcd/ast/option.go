package ast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/pkg/errors"
)

// Option -
type Option struct {
	Default
	Child Node
}

// NewOption -
func NewOption(depth int) *Option {
	return &Option{
		Default: NewDefault(consts.OPTION, 1, depth),
	}
}

// String -
func (opt *Option) String() string {
	var s strings.Builder
	s.WriteString(opt.Default.String())
	s.WriteString(strings.Repeat(base.DefaultIndent, opt.depth))
	s.WriteString(opt.Child.String())
	return s.String()
}

// MarshalJSON -
func (opt *Option) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.OPTION, opt.annots, opt.Child)
}

// ParseType -
func (opt *Option) ParseType(node *base.Node, id *int) error {
	if err := opt.Default.ParseType(node, id); err != nil {
		return err
	}

	child, err := typingNode(node.Args[0], opt.depth, id)
	if err != nil {
		return err
	}
	opt.Child = child

	return nil
}

// ParseValue -
func (opt *Option) ParseValue(node *base.Node) error {
	if len(node.Args) > opt.argsCount {
		return errors.Wrap(base.ErrTreesAreDifferent, "Option.ParseValue")
	}

	switch node.Prim {
	case consts.None:
		opt.Value = nil
		return nil
	case consts.Some:
		err := opt.Child.ParseValue(node.Args[0])
		return err
	default:
		return base.ErrInvalidPrim
	}

}

// ToMiguel -
func (opt *Option) ToMiguel() (*MiguelNode, error) {
	var ast Node

	if opt.Value == nil {
		ast = &opt.Default
	} else {
		ast = opt.Child
	}

	node, err := ast.ToMiguel()
	if err != nil {
		return nil, err
	}
	node.IsOption = true
	return node, nil
}
