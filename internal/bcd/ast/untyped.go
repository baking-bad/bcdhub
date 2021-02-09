package ast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
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
		type buf UntypedAST
		return json.Unmarshal(data, (*buf)(u))
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
		node, err := typingNode(u[i], 0, &id)
		if err != nil {
			return ast, err
		}
		ast.Nodes = append(ast.Nodes, node)
	}
	return ast, nil
}

func typingNode(node *base.Node, depth int, id *int) (Node, error) {
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
	case consts.STORAGE:
		ast = NewStorage(depth + 1)
	case consts.CODE:
		ast = NewCode(depth + 1)
	case consts.BLS12381FR:
		ast = NewBLS12381fr(depth + 1)
	case consts.BLS12381G1:
		ast = NewBLS12381g1(depth + 1)
	case consts.BLS12381G2:
		ast = NewBLS12381g2(depth + 1)
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
