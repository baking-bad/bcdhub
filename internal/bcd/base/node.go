package base

import (
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Node - struct for parsing micheline
type Node struct {
	Prim        string        `json:"prim,omitempty"`
	Args        []*Node       `json:"args,omitempty"`
	Annots      []string      `json:"annots,omitempty"`
	StringValue *string       `json:"string,omitempty"`
	BytesValue  *string       `json:"bytes,omitempty"`
	IntValue    *types.BigInt `json:"int,omitempty"`
}

// MaxNodeDepth - how deep Micheline nodes may nest in JSON. It is at least
// the depth the indexer accepted before (about 10000 nested Pair) and keeps
// recursion bounded on API input.
const MaxNodeDepth = 1 << 14

// UnmarshalJSON -
//
// The whole tree is decoded in a single pass over data. Delegating nested
// nodes back to json.Unmarshal made the decoder re-scan and copy each subtree
// once per nesting level, which is quadratic on deeply nested input.
func (node *Node) UnmarshalJSON(data []byte) error {
	iter := json.BorrowIterator(data)
	defer json.ReturnIterator(iter)

	if err := node.decode(iter, 1); err != nil {
		return err
	}
	if iter.Error != nil {
		return iter.Error
	}
	// the input must end after the node: only whitespace may follow
	iter.WhatIsNext()
	if iter.Error != io.EOF {
		return errors.Wrap(consts.ErrInvalidJSON, "unexpected data after the node")
	}
	return nil
}

func (node *Node) decode(iter *jsoniter.Iterator, depth int) error {
	if depth > MaxNodeDepth {
		return errors.Wrapf(consts.ErrInvalidJSON, "nodes are nested deeper than %d", MaxNodeDepth)
	}

	switch iter.WhatIsNext() {
	case jsoniter.ArrayValue:
		node.Prim = consts.PrimArray
		node.Args = make([]*Node, 0)
		return decodeNodes(iter, &node.Args, depth)
	case jsoniter.ObjectValue:
		for field := iter.ReadObject(); field != ""; field = iter.ReadObject() {
			if err := node.decodeField(iter, strings.ToLower(field), depth); err != nil {
				return err
			}
			if iter.Error != nil {
				return iter.Error
			}
		}
		return iter.Error
	default:
		return consts.ErrInvalidJSON
	}
}

func (node *Node) decodeField(iter *jsoniter.Iterator, field string, depth int) error {
	if iter.WhatIsNext() == jsoniter.NilValue {
		// like encoding/json: null resets pointer and slice fields
		node.resetField(field)
		iter.Skip()
		return nil
	}
	switch field {
	case "prim":
		node.Prim = iter.ReadString()
	case "args":
		if iter.WhatIsNext() != jsoniter.ArrayValue {
			return consts.ErrInvalidJSON
		}
		node.Args = make([]*Node, 0)
		return decodeNodes(iter, &node.Args, depth)
	case "annots":
		node.Annots = make([]string, 0)
		for iter.ReadArray() {
			node.Annots = append(node.Annots, iter.ReadString())
			if iter.Error != nil {
				return iter.Error
			}
		}
	case "string":
		value := iter.ReadString()
		node.StringValue = &value
	case "bytes":
		value := iter.ReadString()
		node.BytesValue = &value
	case "int":
		node.IntValue = new(types.BigInt)
		return node.IntValue.UnmarshalJSON(iter.SkipAndReturnBytes())
	default:
		iter.Skip()
	}
	return nil
}

func (node *Node) resetField(field string) {
	switch field {
	case "args":
		node.Args = nil
	case "annots":
		node.Annots = nil
	case "string":
		node.StringValue = nil
	case "bytes":
		node.BytesValue = nil
	case "int":
		node.IntValue = nil
	}
}

func decodeNodes(iter *jsoniter.Iterator, nodes *[]*Node, depth int) error {
	for iter.ReadArray() {
		if iter.WhatIsNext() == jsoniter.NilValue {
			iter.Skip()
			*nodes = append(*nodes, nil)
			continue
		}
		child := new(Node)
		if err := child.decode(iter, depth+1); err != nil {
			return err
		}
		*nodes = append(*nodes, child)
		if iter.Error != nil {
			return iter.Error
		}
	}
	return iter.Error
}

// MarshalJSON -
func (node *Node) MarshalJSON() ([]byte, error) {
	if node.Prim == consts.PrimArray {
		return json.Marshal(node.Args)
	}

	type buf Node
	return json.Marshal((*buf)(node))
}

// GetAnnotations - returns all node`s annotations recursively
func (node *Node) GetAnnotations() map[string]struct{} {
	annots := make(map[string]struct{})
	for i := range node.Annots {
		if len(node.Annots[i]) == 0 {
			continue
		}
		if node.Annots[i][0] == consts.AnnotPrefixFieldName || node.Annots[i][0] == consts.AnnotPrefixrefixTypeName {
			annots[node.Annots[i][1:]] = struct{}{}
		}
	}
	for i := range node.Args {
		for k := range node.Args[i].GetAnnotations() {
			annots[k] = struct{}{}
		}
	}
	return annots
}

// Compare -
func (node *Node) Compare(second *Node) bool {
	if node.Prim != second.Prim {
		return false
	}
	if len(node.Args) != len(second.Args) {
		return false
	}
	for i := range node.Args {
		if !node.Args[i].Compare(second.Args[i]) {
			return false
		}
	}
	return true
}

// String - converts node info to string
func (node *Node) String() string {
	return node.print(0) + "\n"
}

func (node *Node) print(depth int) string {
	var s strings.Builder
	s.WriteByte('\n')
	s.WriteString(strings.Repeat(consts.DefaultIndent, depth))
	switch {
	case node.Prim != "":
		s.WriteString(node.Prim)
		for i := range node.Args {
			s.WriteString(node.Args[i].print(depth + 1))
		}
	case node.IntValue != nil:
		fmt.Fprintf(&s, "Int=%d", *node.IntValue)
	case node.BytesValue != nil:
		fmt.Fprintf(&s, "Bytes=%s", *node.BytesValue)
	case node.StringValue != nil:
		fmt.Fprintf(&s, "String=%s", *node.StringValue)
	}
	return s.String()
}

var lambdaPackedReg = regexp.MustCompile("^0502[0-9a-f]{8}0[3-9]")

// IsLambda -
func (node *Node) IsLambda() bool {
	if node.BytesValue == nil {
		return false
	}
	input := *node.BytesValue
	if len(input) < 24 {
		return false
	}
	if !lambdaPackedReg.MatchString(input) {
		return false
	}
	b, err := hex.DecodeString(input[22:24])
	if err != nil {
		log.Err(err).Str("input", input).Msg("decoding lambda")
		return false
	}
	if len(b) != 1 {
		return false
	}
	if 0x0c > b[0] || 0x75 < b[0] {
		return false
	}
	return 0x58 >= b[0] || 0x6f <= b[0]

}

// IsComparable reports whether the node, read as a Michelson type expression,
// is a comparable type.
//
// Mirrors the protocol: comparability of each type is declared in
// lib_protocol/script_typed_ir.ml (is_comparable), and the comparable parser
// is parse_comparable_ty in lib_protocol/script_ir_translator.ml. Additionally:
//   - tx_rollup_l2_address was comparable in the protocols that had tx rollups;
//   - unexpanded global constants are accepted: they cannot be checked here.
func (node *Node) IsComparable() bool {
	if node == nil {
		return false
	}
	switch strings.ToLower(node.Prim) {
	case consts.UNIT, consts.NEVER, consts.BOOL, consts.INT, consts.NAT, consts.STRING,
		consts.CHAINID, consts.BYTES, consts.MUTEZ, consts.KEYHASH, consts.KEY,
		consts.SIGNATURE, consts.TIMESTAMP, consts.ADDRESS, consts.TXROLLUPL2ADDRESS,
		consts.CONSTANT:
		return true
	case consts.OPTION, consts.OR, consts.PAIR:
		if len(node.Args) == 0 {
			return false
		}
		for i := range node.Args {
			if !node.Args[i].IsComparable() {
				return false
			}
		}
		return true
	default:
		return false
	}
}
