package ast

import (
	"encoding/hex"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	"github.com/pkg/errors"
)

// UntypedAST -
type UntypedAST []*base.Node

// NewUntypedAST -
func NewUntypedAST(data []byte) (UntypedAST, error) {
	ast := make(UntypedAST, 0)
	err := json.Unmarshal(data, &ast)
	return ast, err
}

// String -
func (u UntypedAST) String() string {
	var s strings.Builder

	for i := range u {
		s.WriteString(u[i].String())
	}

	return s.String()
}

// Hash -
func (u UntypedAST) Hash() (string, error) {
	var s strings.Builder

	for i := range u {
		h, err := u[i].Hash()
		if err != nil {
			return "", err
		}
		s.WriteString(h)
	}
	return s.String(), nil
}

// Annotations -
func (u UntypedAST) Annotations() []string {
	annots := make([]string, 0)
	for i := range u {
		for k := range u[i].GetAnnotations() {
			annots = append(annots, k)
		}
	}
	return annots
}

// UnmarshalJSON -
func (u *UntypedAST) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return consts.ErrInvalidJSON
	}
	if data[0] == '[' {
		node := base.Node{
			Prim: consts.PrimArray,
		}
		var args []*base.Node
		if err := json.Unmarshal(data, &args); err != nil {
			return err
		}
		node.Args = args
		*u = append(*u, &node)
		return nil
	} else if data[0] == '{' {
		var node base.Node
		if err := json.Unmarshal(data, &node); err != nil {
			return err
		}
		*u = append(*u, &node)
		return nil
	}
	return consts.ErrInvalidJSON
}

// ToTypedAST -
func (u UntypedAST) ToTypedAST() (*TypedAst, error) {
	ast := NewTypedAST()
	id := 0
	for i := range u {
		if u[i].Prim == consts.PrimArray {
			for j := range u[i].Args {
				node, err := typeNode(u[i].Args[j], 0, &id)
				if err != nil {
					return ast, err
				}
				ast.Nodes = append(ast.Nodes, node)
			}
		} else {
			node, err := typeNode(u[i], 0, &id)
			if err != nil {
				return ast, err
			}
			ast.Nodes = append(ast.Nodes, node)
		}
	}
	return ast, nil
}

// GetStrings -
func (u UntypedAST) GetStrings(tryUnpack bool) ([]string, error) {
	s := make([]string, 0)
	for i := range u {
		arr, err := forge.CollectStrings(u[i], tryUnpack)
		if err != nil {
			return nil, err
		}
		s = append(s, arr...)
	}
	return s, nil
}

// Fingerprint -
func (u UntypedAST) Fingerprint(isCode bool) ([]byte, error) {
	var s strings.Builder
	for i := range u {
		f, err := u[i].Fingerprint(isCode)
		if err != nil {
			return nil, err
		}
		if _, err := s.WriteString(f); err != nil {
			return nil, err
		}
	}
	return hex.DecodeString(s.String())
}

// Unpack - unpack all bytes and store data in the tree.
func (u UntypedAST) Unpack() {
	for i := range u {
		if u[i].BytesValue == nil {
			continue
		}

		value := *u[i].BytesValue
		tree := forge.TryUnpackString(value)
		if len(tree) > 0 {
			u[i] = tree[0]
		}
	}
}

// Stringify - make readable string
func (u UntypedAST) Stringify() (string, error) {
	switch {
	case len(u) == 1 && len(u[0].Args) > 0:
		str, err := json.MarshalToString(u[0])
		if err != nil {
			return "", err
		}
		return formatter.MichelineToMichelsonInline(str)
	case len(u) == 1 && u[0].StringValue != nil:
		return *u[0].StringValue, nil
	case len(u) == 1 && u[0].IntValue != nil:
		return u[0].IntValue.String(), nil
	case len(u) == 1 && u[0].BytesValue != nil:
		tree := forge.TryUnpackString(*u[0].BytesValue)
		if tree == nil {
			return *u[0].BytesValue, nil
		}
		treeJSON, err := json.MarshalToString(tree)
		if err != nil {
			return "", err
		}
		return formatter.MichelineToMichelsonInline(treeJSON)
	default:
		str, err := json.MarshalToString(u)
		if err != nil {
			return "", err
		}
		return formatter.MichelineToMichelsonInline(str)
	}
}

func typeNode(node *base.Node, depth int, id *int) (Node, error) {
	var ast Node
	switch strings.ToLower(node.Prim) {
	case consts.UNIT:
		ast = NewUnit(depth + 1)
	case consts.STRING:
		ast = NewString(depth + 1)
	case consts.INT:
		ast = NewInt(depth + 1)
	case consts.NAT:
		ast = NewNat(depth + 1)
	case consts.MUTEZ:
		ast = NewMutez(depth + 1)
	case consts.BOOL:
		ast = NewBool(depth + 1)
	case consts.TIMESTAMP:
		ast = NewTimestamp(depth + 1)
	case consts.BYTES:
		ast = NewBytes(depth + 1)
	case consts.NEVER:
		ast = NewNever(depth + 1)
	case consts.OPERATION:
		ast = NewOperation(depth + 1)
	case consts.CHAINID:
		ast = NewChainID(depth + 1)
	case consts.ADDRESS:
		ast = NewAddress(depth + 1)
	case consts.KEY:
		ast = NewKey(depth + 1)
	case consts.KEYHASH:
		ast = NewKeyHash(depth + 1)
	case consts.SIGNATURE:
		ast = NewSignature(depth + 1)
	case consts.BIGMAP:
		ast = NewBigMap(depth + 1)
	case consts.CONTRACT:
		ast = NewContract(depth + 1)
	case consts.LAMBDA:
		ast = NewLambda(depth + 1)
	case consts.LIST:
		ast = NewList(depth + 1)
	case consts.MAP:
		ast = NewMap(depth + 1)
	case consts.OPTION:
		ast = NewOption(depth + 1)
	case consts.OR:
		ast = NewOr(depth + 1)
	case consts.PAIR:
		ast = NewPair(depth + 1)
	case consts.SAPLINGSTATE:
		ast = NewSaplingState(depth + 1)
	case consts.SAPLINGTRANSACTION:
		ast = NewSaplingTransaction(depth + 1)
	case consts.SET:
		ast = NewSet(depth + 1)
	case consts.TICKET:
		ast = NewTicket(depth + 1)
	case consts.PARAMETER:
		ast = NewParameter(depth + 1)
	case consts.BLS12381FR:
		ast = NewBLS12381fr(depth + 1)
	case consts.BLS12381G1:
		ast = NewBLS12381g1(depth + 1)
	case consts.BLS12381G2:
		ast = NewBLS12381g2(depth + 1)
	case consts.BAKERHASH:
		ast = NewBakerHash(depth + 1)
	case consts.CHEST:
		ast = NewChest(depth + 1)
	case consts.CHESTKEY:
		ast = NewChestKey(depth + 1)
	default:
		return nil, errors.Wrap(consts.ErrUnknownPrim, node.Prim)
	}

	return ast, ast.ParseType(node, id)
}

func getAnnotation(x []string, prefix byte) string {
	for i := range x {
		if x[i][0] == prefix {
			return x[i][1:]
		}
	}
	return ""
}
