package ast

import "github.com/baking-bad/bcdhub/internal/bcd/consts"

// Code -
type Code struct {
	Default

	depth int
}

// NewCode -
func NewCode(depth int) *Code {
	return &Code{
		Default: NewDefault(consts.CODE, -1, depth),
	}
}
