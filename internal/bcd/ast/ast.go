package ast

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	jsoniter "github.com/json-iterator/go"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
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

// Forge -
func (a *TypedAst) Forge(optimized bool) ([]byte, error) {
	data := make([]byte, 0)
	for i := range a.Nodes {
		nodeData, err := a.Nodes[i].Forge(optimized)
		if err != nil {
			return data, err
		}
		data = append(data, nodeData...)
	}
	return data, nil
}

// Unforge -
func (a *TypedAst) Unforge(data []byte) (int, error) {
	var count int
	for i := range a.Nodes {
		n, err := a.Nodes[i].Unforge(data[count:])
		if err != nil {
			return count, err
		}
		count += n
	}
	return count, nil
}

// Pack -
func (a *TypedAst) Pack() ([]byte, error) {
	data := make([]byte, 0)
	for i := range a.Nodes {
		args, err := a.Nodes[i].Pack()
		if err != nil {
			return nil, err
		}
		data = append(data, args...)
	}
	return data, nil
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
