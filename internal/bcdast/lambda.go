package bcdast

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
)

// Lambda -
type Lambda struct {
	Default
	Type AstNode
}

// NewLambda -
func NewLambda(depth int) *Lambda {
	return &Lambda{
		Default: NewDefault(consts.LAMBDA, 0, depth),
	}
}

// MarshalJSON -
func (l *Lambda) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.LAMBDA, l.annots, l.Type)
}

// ParseType -
func (l *Lambda) ParseType(untyped Untyped, id *int) error {
	if err := l.Default.ParseType(untyped, id); err != nil {
		return err
	}

	typ, err := typingNode(untyped.Args[0], l.depth, id)
	if err != nil {
		return err
	}
	l.Type = typ
	return nil
}

// ParseValue -
func (l *Lambda) ParseValue(untyped Untyped) error {
	return l.Default.ParseValue(untyped)
}
