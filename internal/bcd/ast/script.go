package ast

import (
	"bytes"
	stdJSON "encoding/json"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/pkg/errors"
)

// Script -
type Script struct {
	Code      UntypedAST   `json:"code"`
	Parameter UntypedAST   `json:"parameter"`
	Storage   UntypedAST   `json:"storage"`
	Views     []UntypedAST `json:"-"`
}

// View -
type View struct {
	Name string
	Code UntypedAST
}

type sectionNode struct {
	Prim string             `json:"prim"`
	Args stdJSON.RawMessage `json:"args"`
}

// UnmarshalJSON -
func (s *Script) UnmarshalJSON(data []byte) error {
	var ast UntypedAST
	if err := json.Unmarshal(data, &ast); err != nil {
		return err
	}
	for len(ast) == 1 && ast[0].Prim == consts.PrimArray {
		ast = ast[0].Args
	}
	for i := range ast {
		tree := UntypedAST(ast[i].Args)
		switch ast[i].Prim {
		case consts.PARAMETER:
			s.Parameter = tree
		case consts.STORAGE:
			s.Storage = tree
		case consts.CODE:
			s.Code = tree
		case consts.View:
			s.Views = append(s.Views, tree)
		default:
			return errors.Wrap(consts.ErrUnknownPrim, fmt.Sprintf("NewScript : %s", ast[i].Prim))
		}
	}
	return nil
}

// MarshalJSON -
func (s *Script) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(`[{"prim":"parameter","args":`)

	parameter, err := json.Marshal(s.Parameter)
	if err != nil {
		return nil, err
	}
	buf.Write(parameter)

	buf.WriteString(`},{"prim":"storage","args":`)
	storage, err := json.Marshal(s.Storage)
	if err != nil {
		return nil, err
	}
	buf.Write(storage)

	buf.WriteString(`},{"prim":"code","args":`)
	code, err := json.Marshal(s.Code)
	if err != nil {
		return nil, err
	}
	buf.Write(code)

	for i := range s.Views {
		buf.WriteString(`},{"prim":"view","args":`)
		view, err := json.Marshal(s.Views[i])
		if err != nil {
			return nil, err
		}
		buf.Write(view)
	}
	buf.WriteString("}]")

	return buf.Bytes(), nil
}

// NewScript - creates `Script` object: untyped trees of code, storage and parameter
func NewScript(data []byte) (*Script, error) {
	var s Script
	err := json.Unmarshal(data, &s)
	return &s, err
}

// NewScriptWithoutCode - creates `Script` object without code tree: storage and parameter
func NewScriptWithoutCode(data []byte) (*Script, error) {
	var sections []sectionNode
	if err := json.Unmarshal(data, &sections); err != nil {
		return nil, err
	}

	var script Script
	for i := range sections {
		switch sections[i].Prim {
		case consts.PARAMETER:
			var tree UntypedAST
			if err := json.Unmarshal(sections[i].Args, &tree); err != nil {
				return nil, err
			}
			script.Parameter = tree
		case consts.STORAGE:
			var tree UntypedAST
			if err := json.Unmarshal(sections[i].Args, &tree); err != nil {
				return nil, err
			}
			script.Storage = tree
		}
	}

	return &script, nil
}

// Compare - compares two scripts
func (s *Script) Compare(another *Script) bool {
	if len(s.Parameter) != len(another.Parameter) {
		return false
	}
	for i := range s.Parameter {
		if !s.Parameter[i].Compare(another.Parameter[i]) {
			return false
		}
	}
	if len(s.Storage) != len(another.Storage) {
		return false
	}
	for i := range s.Storage {
		if !s.Storage[i].Compare(another.Storage[i]) {
			return false
		}
	}

	if len(s.Code) != len(another.Code) {
		return false
	}
	for i := range s.Code {
		if !s.Code[i].Compare(another.Code[i]) {
			return false
		}
	}
	return true
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
		child, err := typeNode(arg, st.depth, id)
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
