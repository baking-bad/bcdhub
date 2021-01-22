package bcdast

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/pkg/errors"
)

// Script -
type Script struct {
	Code      UntypedAST
	Parameter UntypedAST
	Storage   UntypedAST
}

// NewScript -
func NewScript(data []byte) (*Script, error) {
	var ast UntypedAST
	if err := json.Unmarshal(data, &ast); err != nil {
		return nil, err
	}
	var s Script
	for i := range ast {
		tree := UntypedAST(ast[i].Args)
		switch ast[i].Prim {
		case consts.PARAMETER:
			s.Parameter = tree
		case consts.STORAGE:
			s.Storage = tree
		case consts.CODE:
			s.Code = tree
		default:
			return nil, errors.Wrap(ErrUnknownPrim, fmt.Sprintf("NewScript : %s", ast[i].Prim))
		}
	}
	return &s, nil
}

// SectionType -
type SectionType struct {
	Default

	Args []AstNode

	depth int
}

// NewSectionType -
func NewSectionType(prim string, depth int) *SectionType {
	return &SectionType{
		Default: NewDefault(prim, -1, depth),
	}
}

// String -
func (st *SectionType) String() string {
	var s strings.Builder
	s.WriteString(st.Default.String())
	for i := range st.Args {
		s.WriteString(strings.Repeat(indent, st.depth))
		s.WriteString(st.Args[i].String())
	}
	return s.String()
}

// MarshalJSON -
func (st *SectionType) MarshalJSON() ([]byte, error) {
	return marshalJSON(st.Prim, st.annots, st.Args...)
}

// ParseType -
func (st *SectionType) ParseType(untyped Untyped, id *int) error {
	if err := st.Default.ParseType(untyped, id); err != nil {
		return err
	}

	st.Args = make([]AstNode, 0, len(untyped.Args))
	for _, arg := range untyped.Args {
		child, err := typingNode(arg, st.depth, id)
		if err != nil {
			return err
		}
		st.Args = append(st.Args, child)
	}

	return nil
}

// ParseValue -
func (st *SectionType) ParseValue(untyped Untyped) error {
	for i := range untyped.Args {
		if err := st.Args[0].ParseValue(untyped.Args[i]); err != nil {
			return err
		}
	}
	return nil
}
