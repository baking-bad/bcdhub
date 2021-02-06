package ast

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	jsoniter "github.com/json-iterator/go"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/pkg/errors"
	"github.com/ulule/deepcopier"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// TypedAst -
type TypedAst struct {
	Nodes   []Node
	settled bool
}

// NewTypedAST -
func NewTypedAST() *TypedAst {
	return &TypedAst{
		Nodes: make([]Node, 0),
	}
}

// IsSettled -
func (a *TypedAst) IsSettled() bool {
	return a.settled
}

// String -
func (a *TypedAst) String() string {
	var s strings.Builder
	for i := range a.Nodes {
		s.WriteString(a.Nodes[i].String())
	}
	return s.String()
}

// Settle -
func (a *TypedAst) Settle(untyped UntypedAST) error {
	if len(untyped) != len(a.Nodes) && len(a.Nodes) == 1 {
		if _, ok := a.Nodes[0].(*Pair); ok {
			newUntyped := &base.Node{
				Prim: consts.Pair,
				Args: untyped,
			}
			if err := a.Nodes[0].ParseValue(newUntyped); err != nil {
				return err
			}
			a.settled = true
			return nil
		}
	} else if len(untyped) == len(a.Nodes) {
		for i := range untyped {
			if err := a.Nodes[i].ParseValue(untyped[i]); err != nil {
				return err
			}
		}
		a.settled = true
		return nil
	}
	return errors.Wrap(base.ErrTreesAreDifferent, "TypedAst.MergeValue")
}

// ToMiguel -
func (a *TypedAst) ToMiguel() ([]*MiguelNode, error) {
	nodes := make([]*MiguelNode, 0)
	for i := range a.Nodes {
		m, err := a.Nodes[i].ToMiguel()
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, m)
	}
	return nodes, nil
}

// GetEntrypoints -
func (a *TypedAst) GetEntrypoints() []string {
	entrypoints := make([]string, 0)
	for i := range a.Nodes {
		entrypoints = append(entrypoints, a.Nodes[i].GetEntrypoints()...)
	}
	if len(entrypoints) == 1 && entrypoints[0] == "" {
		entrypoints[0] = consts.DefaultEntrypoint
	} else {
		for i := range entrypoints {
			if entrypoints[i] == "" {
				entrypoints[i] = fmt.Sprintf("entrypoint_%d", i)
			}
		}
	}

	return entrypoints
}

// ToBaseNode -
func (a *TypedAst) ToBaseNode(optimized bool) (*base.Node, error) {
	if len(a.Nodes) == 1 {
		return a.Nodes[0].ToBaseNode(optimized)
	}
	return arrayToBaseNode(a.Nodes, optimized)
}

// ToJSONSchema -
func (a *TypedAst) ToJSONSchema() (*JSONSchema, error) {
	if len(a.Nodes) == 1 {
		if a.Nodes[0].GetPrim() == consts.UNIT {
			return nil, nil
		}
		return a.Nodes[0].ToJSONSchema()
	}

	s := &JSONSchema{
		Type:       JSONSchemaTypeObject,
		Properties: make(map[string]*JSONSchema),
	}

	for i := range a.Nodes {
		if err := setChildSchema(a.Nodes[i], false, s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// FromJSONSchema -
func (a *TypedAst) FromJSONSchema(data map[string]interface{}) error {
	for i := range a.Nodes {
		if err := a.Nodes[i].FromJSONSchema(data); err != nil {
			return err
		}
	}
	return nil
}

// ToParameters -
func (a *TypedAst) ToParameters() ([]byte, error) {
	if len(a.Nodes) == 1 {
		return a.Nodes[0].ToParameters()
	}

	return buildListParameters(a.Nodes)
}

func createByType(typ Node) (Node, error) {
	obj := reflect.New(reflect.ValueOf(typ).Elem().Type()).Interface().(Node)
	return obj, deepcopier.Copy(typ).To(obj)
}

func marshalJSON(prim string, annots []string, args ...Node) ([]byte, error) {
	var builder bytes.Buffer
	builder.WriteByte('{')
	builder.WriteString(fmt.Sprintf(`"prim": "%s"`, prim))
	if len(args) > 0 {
		builder.WriteString(`, "args": [`)
		for i := range args {
			typ, err := json.Marshal(args[i])
			if err != nil {
				return nil, err
			}
			if _, err := builder.Write(typ); err != nil {
				return nil, err
			}
			if i < len(args)-1 {
				builder.WriteByte(',')
			}
		}
		builder.WriteByte(']')
	}
	if len(annots) > 0 {
		builder.WriteString(fmt.Sprintf(`, "annots": ["%s"]`, strings.Join(annots, `","`)))

	}
	builder.WriteByte('}')
	return builder.Bytes(), nil
}
