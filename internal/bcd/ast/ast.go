package ast

import (
	"bytes"
	"fmt"
	"strings"

	jsoniter "github.com/json-iterator/go"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/pkg/errors"
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

// NewTypedAstFromBytes -
func NewTypedAstFromBytes(data []byte) (*TypedAst, error) {
	var tree UntypedAST
	if err := json.Unmarshal(data, &tree); err != nil {
		return nil, err
	}
	return tree.ToTypedAST()
}

// NewTypedAstFromString -
func NewTypedAstFromString(data string) (*TypedAst, error) {
	var tree UntypedAST
	if err := json.UnmarshalFromString(data, &tree); err != nil {
		return nil, err
	}
	return tree.ToTypedAST()
}

// NewSettledTypedAst -
func NewSettledTypedAst(tree, data string) (*TypedAst, error) {
	typ, err := NewTypedAstFromString(tree)
	if err != nil {
		return nil, err
	}

	var treeData UntypedAST
	if err := json.UnmarshalFromString(data, &treeData); err != nil {
		return nil, err
	}

	err = typ.Settle(treeData)
	return typ, err
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
	return errors.Wrap(consts.ErrTreesAreDifferent, "TypedAst.Settle")
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
		return []string{consts.DefaultEntrypoint}
	}

	for i := range entrypoints {
		if entrypoints[i] == "" {
			entrypoints[i] = fmt.Sprintf("entrypoint_%d", i)
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

// FindByName -
func (a *TypedAst) FindByName(name string, isEntrypoint bool) Node {
	for i := range a.Nodes {
		if node := a.Nodes[i].FindByName(name, isEntrypoint); node != nil {
			return node
		}
	}
	return nil
}

// ToParameters -
func (a *TypedAst) ToParameters(entrypoint string) ([]byte, error) {
	if entrypoint == "" {
		if len(a.Nodes) == 1 {
			return a.Nodes[0].ToParameters()
		}

		return buildListParameters(a.Nodes)
	}

	node := a.FindByName(entrypoint, true)
	if node != nil {
		return node.ToParameters()
	}
	return nil, nil
}

// Docs -
func (a *TypedAst) Docs(entrypoint string) ([]Typedef, error) {
	if entrypoint == DocsFull {
		if len(a.Nodes) == 1 {
			docs, _, err := a.Nodes[0].Docs(DocsFull)
			return docs, err
		}
		return buildArrayDocs(a.Nodes)
	}

	node := a.FindByName(entrypoint, true)
	if node != nil {
		docs, _, err := node.Docs(DocsFull)
		return docs, err
	}
	return nil, nil
}

// GetEntrypointsDocs -
func (a *TypedAst) GetEntrypointsDocs() ([]EntrypointType, error) {
	docs, err := a.Docs(DocsFull)
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, nil
	}

	if docs[0].Type == consts.OR {
		response := make([]EntrypointType, 0)
		for i := range docs[0].Args {
			name := docs[0].Args[i].Key
			if strings.HasPrefix(name, "@") {
				name = fmt.Sprintf("entrypoint_%d", len(response))
			}
			entrypoint := EntrypointType{
				Name: name,
			}
			eDocs, err := a.Docs(docs[0].Args[i].Key)
			if err != nil {
				return nil, err
			}
			entrypoint.Type = eDocs
			response = append(response, entrypoint)
		}

		return response, nil
	}
	entrypoint := EntrypointType{
		Name: consts.DefaultEntrypoint,
		Type: docs,
	}
	return []EntrypointType{entrypoint}, nil
}

// Compare -
func (a *TypedAst) Compare(b *TypedAst) (int, error) {
	if len(a.Nodes) != len(b.Nodes) {
		return 0, consts.ErrTypeIsNotComparable
	}
	for i := range a.Nodes {
		res, err := a.Nodes[i].Compare(b.Nodes[i])
		if err != nil {
			return res, err
		}
		if res != 0 {
			return res, nil
		}
	}
	return 0, nil
}

// Diff -
func (a *TypedAst) Diff(b *TypedAst) (*MiguelNode, error) {
	if b == nil {
		tree, err := a.ToMiguel()
		if err != nil {
			return nil, err
		}
		if len(tree) == 0 {
			return nil, nil
		}
		for i := range tree {
			tree[i].setDiffType(MiguelKindCreate)
		}
		return tree[0], nil
	}
	if len(b.Nodes) == 1 && len(a.Nodes) == 1 {
		return a.Nodes[0].Distinguish(b.Nodes[0])
	}
	return nil, nil
}

// FromParameters - fill(settle) subtree in `a` with `data`. Returned settled subtree and error if occured. If `entrypoint` is empty string it will try to settle main tree.
func (a *TypedAst) FromParameters(data *types.Parameters) (*TypedAst, error) {
	var tree UntypedAST
	if err := json.Unmarshal(data.Value, &tree); err != nil {
		return nil, err
	}

	if data.Entrypoint != "" {
		subTree := a.FindByName(data.Entrypoint, true)
		if subTree != nil {
			err := subTree.ParseValue(tree[0])
			return &TypedAst{
				Nodes:   []Node{subTree},
				settled: true,
			}, err
		}
	}
	err := a.Settle(tree)
	return a, err
}

// EnrichBigMap -
func (a *TypedAst) EnrichBigMap(bmd []*types.BigMapDiff) error {
	for i := range a.Nodes {
		if err := a.Nodes[i].EnrichBigMap(bmd); err != nil {
			return err
		}
	}
	return nil
}

// EqualType -
func (a *TypedAst) EqualType(b *TypedAst) bool {
	if len(a.Nodes) != len(b.Nodes) {
		return false
	}

	for i := range a.Nodes {
		if !a.Nodes[i].EqualType(b.Nodes[i]) {
			return false
		}
	}

	return true
}

// FindBigMapByPtr -
func (a *TypedAst) FindBigMapByPtr() map[int64]*BigMap {
	res := make(map[int64]*BigMap)
	for i := range a.Nodes {
		if m := a.Nodes[i].FindPointers(); m != nil {
			for p, bm := range m {
				res[p] = bm
			}
		}
	}
	return res
}

// GetJSONModel -
func (a *TypedAst) GetJSONModel(model JSONModel) {
	if model == nil {
		model = make(JSONModel)
	}

	for i := range a.Nodes {
		a.Nodes[i].GetJSONModel(model)
	}
}

// Unwrap - clear paramaters from Left/Right. Tree must be settled.
func (a *TypedAst) Unwrap() Node {
	if !a.IsSettled() || len(a.Nodes) != 1 {
		return nil
	}

	return unwrap(a.Nodes[0])
}

func unwrap(node Node) Node {
	or, ok := node.(*Or)
	if !ok {
		return node
	}

	switch or.key {
	case leftKey:
		return unwrap(or.LeftType)
	case rightKey:
		return unwrap(or.RightType)
	}
	return node
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
