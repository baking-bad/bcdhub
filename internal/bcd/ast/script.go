package ast

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/pkg/errors"
)

// Script -
type Script struct {
	Code      UntypedAST
	Parameter UntypedAST
	Storage   UntypedAST
}

// NewScript - creates `Script` object: untyped trees of code, storage and parameter
func NewScript(data []byte) (*Script, error) {
	var ast UntypedAST
	if err := json.Unmarshal(data, &ast); err != nil {
		return nil, err
	}

	if len(ast) == 1 && ast[0].Prim == consts.PrimArray {
		ast = ast[0].Args
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
			return nil, errors.Wrap(consts.ErrUnknownPrim, fmt.Sprintf("NewScript : %s", ast[i].Prim))
		}
	}
	return &s, nil
}

// StorageType - returns storage`s typed tree
func (s *Script) StorageType() (*TypedAst, error) {
	return s.Storage.ToTypedAST()
}

// ParameterType - returns parameter`s typed tree
func (s *Script) ParameterType() (*TypedAst, error) {
	return s.Parameter.ToTypedAST()
}

// SectionType -
type SectionType struct {
	Default

	Args []Node

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
		s.WriteString(strings.Repeat(consts.DefaultIndent, st.depth))
		s.WriteString(st.Args[i].String())
	}
	return s.String()
}

// MarshalJSON -
func (st *SectionType) MarshalJSON() ([]byte, error) {
	return marshalJSON(st.Prim, st.Annots, st.Args...)
}

// ParseType -
func (st *SectionType) ParseType(node *base.Node, id *int) error {
	if err := st.Default.ParseType(node, id); err != nil {
		return err
	}

	st.Args = make([]Node, 0, len(node.Args))
	for _, arg := range node.Args {
		child, err := typingNode(arg, st.depth, id)
		if err != nil {
			return err
		}
		st.Args = append(st.Args, child)
	}

	return nil
}

// ParseValue -
func (st *SectionType) ParseValue(node *base.Node) error {
	for i := range node.Args {
		if err := st.Args[0].ParseValue(node.Args[i]); err != nil {
			return err
		}
	}
	return nil
}

// EqualType -
func (st *SectionType) EqualType(node Node) bool {
	if !st.Default.EqualType(node) {
		return false
	}
	second, ok := node.(*Parameter)
	if !ok {
		return false
	}

	if len(st.Args) != len(second.Args) {
		return false
	}

	for i := range st.Args {
		if !st.Args[i].EqualType(second.Args[i]) {
			return false
		}
	}

	return true
}
