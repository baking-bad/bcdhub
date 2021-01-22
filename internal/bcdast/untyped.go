package bcdast

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Untyped -
type Untyped struct {
	Prim        string    `json:"prim,omitempty"`
	Args        []Untyped `json:"args,omitempty"`
	Annots      []string  `json:"annots,omitempty"`
	IntValue    *int64    `json:"int,omitempty,string"`
	BytesValue  *string   `json:"bytes,omitempty"`
	StringValue *string   `json:"string,omitempty"`
}

// UnmarshalJSON -
func (u *Untyped) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return ErrInvalidJSON
	}
	if data[0] == '[' {
		u.Prim = PrimArray
		u.Args = make([]Untyped, 0)
		return json.Unmarshal(data, &u.Args)
	} else if data[0] == '{' {
		type buf Untyped
		return json.Unmarshal(data, (*buf)(u))
	}
	return ErrInvalidJSON
}

func (u *Untyped) getAnnotations() map[string]struct{} {
	annots := make(map[string]struct{}, 0)
	for i := range u.Annots {
		if len(u.Annots[i]) == 0 {
			continue
		}
		if u.Annots[i][0] == prefixFieldName || u.Annots[i][0] == prefixTypeName {
			annots[u.Annots[i][1:]] = struct{}{}
		}
	}
	for i := range u.Args {
		for k := range u.Args[i].getAnnotations() {
			annots[k] = struct{}{}
		}
	}
	return annots
}

// Hash -
func (u *Untyped) Hash() (string, error) {
	var s strings.Builder
	var prim string
	switch {
	case u.Prim != "":
		if u.Prim != consts.RENAME && u.Prim != consts.CAST {
			hashCode, err := getHashCode(u.Prim)
			if err != nil {
				return "", err
			}
			s.WriteString(hashCode)
		}

		for i := range u.Args {
			childHashCode, err := u.Args[i].Hash()
			if err != nil {
				return "", err
			}
			s.WriteString(childHashCode)
		}
		return s.String(), nil
	case u.BytesValue != nil:
		prim = consts.BYTES
	case u.IntValue != nil:
		prim = consts.INT
	case u.StringValue != nil:
		prim = consts.STRING
	}
	hashCode, err := getHashCode(prim)
	if err != nil {
		return "", err
	}
	s.WriteString(hashCode)
	return s.String(), nil
}

// String -
func (u *Untyped) String() string {
	return u.print(0) + "\n"
}

func (u *Untyped) print(depth int) string {
	var s strings.Builder
	s.WriteByte('\n')
	s.WriteString(strings.Repeat(indent, depth))
	switch {
	case u.Prim != "":
		s.WriteString(u.Prim)
		for i := range u.Args {
			s.WriteString(u.Args[i].print(depth + 1))
		}
	case u.IntValue != nil:
		s.WriteString(fmt.Sprintf("Int=%d", *u.IntValue))
	case u.BytesValue != nil:
		s.WriteString(fmt.Sprintf("Bytes=%s", *u.BytesValue))
	case u.StringValue != nil:
		s.WriteString(fmt.Sprintf("String=%s", *u.StringValue))
	}
	return s.String()
}

// Unforge -
func (u *Untyped) Unforge(data []byte) error {
	if len(data) != 0 {
		return ErrEmptyUnforgingData
	}
	switch data[0] {
	case 0x00:

	case 0x01:
	case 0x02:
	case 0x03:
	case 0x04:
	case 0x05:
	case 0x06:
	case 0x07:
	case 0x08:
	case 0x09:
	case 0x0a:
	}
	return nil
}

// Forge -
func (u *Untyped) Forge() ([]byte, error) {
	return nil, nil
}

// UntypedAST -
type UntypedAST []Untyped

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
		for k := range u[i].getAnnotations() {
			annots = append(annots, k)
		}
	}
	return annots
}

// UnmarshalJSON -
func (u *UntypedAST) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return ErrInvalidJSON
	}
	if data[0] == '[' {
		type buf UntypedAST
		return json.Unmarshal(data, (*buf)(u))
	} else if data[0] == '{' {
		var untyped Untyped
		if err := json.Unmarshal(data, &untyped); err != nil {
			return err
		}
		*u = append(*u, untyped)
		return nil
	}
	return ErrInvalidJSON
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

func typingNode(untyped Untyped, depth int, id *int) (AstNode, error) {
	var node AstNode
	switch strings.ToLower(untyped.Prim) {
	case consts.UNIT:
		node = NewUnit(depth + 1)
	case consts.STRING:
		node = NewString(depth + 1)
	case consts.INT:
		node = NewInt(depth + 1)
	case consts.NAT:
		node = NewNat(depth + 1)
	case consts.MUTEZ:
		node = NewMutez(depth + 1)
	case consts.BOOL:
		node = NewBool(depth + 1)
	case consts.TIMESTAMP:
		node = NewTimestamp(depth + 1)
	case consts.BYTES:
		node = NewBytes(depth + 1)
	case consts.NEVER:
		node = NewNever(depth + 1)
	case consts.OPERATION:
		node = NewOperation(depth + 1)
	case consts.CHAINID:
		node = NewChainID(depth + 1)
	case consts.ADDRESS:
		node = NewAddress(depth + 1)
	case consts.KEY:
		node = NewKey(depth + 1)
	case consts.KEYHASH:
		node = NewKeyHash(depth + 1)
	case consts.SIGNATURE:
		node = NewSignature(depth + 1)
	case consts.BIGMAP:
		node = NewBigMap(depth + 1)
	case consts.CONTRACT:
		node = NewContract(depth + 1)
	case consts.LAMBDA:
		node = NewLambda(depth + 1)
	case consts.LIST:
		node = NewList(depth + 1)
	case consts.MAP:
		node = NewMap(depth + 1)
	case consts.OPTION:
		node = NewOption(depth + 1)
	case consts.OR:
		node = NewOr(depth + 1)
	case consts.PAIR:
		node = NewPair(depth + 1)
	case consts.SAPLINGSTATE:
		node = NewSaplingState(depth + 1)
	case consts.SAPLINGTRANSACTION:
		node = NewSaplingTransaction(depth + 1)
	case consts.SET:
		node = NewSet(depth + 1)
	case consts.TICKET:
		node = NewTicket(depth + 1)
	case consts.PARAMETER:
		node = NewParameter(depth + 1)
	case consts.STORAGE:
		node = NewStorage(depth + 1)
	case consts.CODE:
		node = NewCode(depth + 1)
	case consts.BLS12381FR:
		node = NewBLS12381fr(depth + 1)
	case consts.BLS12381G1:
		node = NewBLS12381g1(depth + 1)
	case consts.BLS12381G2:
		node = NewBLS12381g2(depth + 1)
	default:
		return nil, errors.Wrap(ErrUnknownPrim, untyped.Prim)
	}

	return node, node.ParseType(untyped, id)
}

func getAnnotation(x []string, prefix byte) string {
	for i := range x {
		if x[i][0] == prefix {
			return x[i][1:]
		}
	}
	return ""
}
