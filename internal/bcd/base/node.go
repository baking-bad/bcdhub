package base

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/logger"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
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

// UnmarshalJSON -
func (node *Node) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return consts.ErrInvalidJSON
	}
	if data[0] == '[' {
		node.Prim = consts.PrimArray
		node.Args = make([]*Node, 0)
		return json.Unmarshal(data, &node.Args)
	} else if data[0] == '{' {
		type buf Node
		return json.Unmarshal(data, (*buf)(node))
	}
	return consts.ErrInvalidJSON
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

// Hash -
func (node *Node) Hash() (string, error) {
	var s strings.Builder
	var prim string
	switch {
	case node.Prim != "":
		if node.Prim != consts.RENAME && node.Prim != consts.CAST && node.Prim != consts.PrimArray {
			hashCode, err := getHashCode(node.Prim)
			if err != nil {
				return "", err
			}
			s.WriteString(hashCode)
		}

		for i := range node.Args {
			childHashCode, err := node.Args[i].Hash()
			if err != nil {
				return "", err
			}
			s.WriteString(childHashCode)
		}
		return s.String(), nil
	case node.BytesValue != nil:
		prim = consts.BYTES
	case node.IntValue != nil:
		prim = consts.INT
	case node.StringValue != nil:
		prim = consts.STRING
	}
	hashCode, err := getHashCode(prim)
	if err != nil {
		return "", err
	}
	s.WriteString(hashCode)
	return s.String(), nil
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
		s.WriteString(fmt.Sprintf("Int=%d", *node.IntValue))
	case node.BytesValue != nil:
		s.WriteString(fmt.Sprintf("Bytes=%s", *node.BytesValue))
	case node.StringValue != nil:
		s.WriteString(fmt.Sprintf("String=%s", *node.StringValue))
	}
	return s.String()
}

// IsLambda -
func (node *Node) IsLambda() bool {
	if node.BytesValue == nil {
		return false
	}
	input := *node.BytesValue
	if len(input) < 24 {
		return false
	}
	re := regexp.MustCompile("^0502[0-9a-f]{8}0[3-9]")
	if !re.MatchString(input) {
		return false
	}
	b, err := hex.DecodeString(input[22:24])
	if err != nil {
		logger.Error(err)
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

// Fingerprint -
func (node *Node) Fingerprint(isCode bool) (string, error) {
	var fgpt strings.Builder
	switch node.Prim {
	case consts.PrimArray:
		for i := range node.Args {
			buf, err := node.Args[i].Fingerprint(isCode)
			if err != nil {
				return "", err
			}
			fgpt.WriteString(buf)
		}
	default:
		if node.Prim != "" {
			if skip(node.Prim, isCode) {
				return "", nil
			}

			if !pass(node.Prim, isCode) {
				code, err := getCode(node.Prim)
				if err != nil {
					return "", err
				}
				fgpt.WriteString(code)
			}

			for i := range node.Args {
				itemFgpt, err := node.Args[i].Fingerprint(isCode)
				if err != nil {
					return "", err
				}
				fgpt.WriteString(itemFgpt)
			}

		} else {
			var prim string
			switch {
			case node.StringValue != nil:
				prim = consts.STRING
			case node.BytesValue != nil:
				prim = consts.BYTES
			case node.IntValue != nil:
				prim = consts.INT
			}
			if prim != "" {
				code, err := getCode(prim)
				if err != nil {
					return "", err
				}
				fgpt.WriteString(code)
			}
		}
	}

	return fgpt.String(), nil
}

func skip(prim string, isCode bool) bool {
	p := strings.ToLower(prim)
	return isCode && (p == consts.CAST || p == consts.RENAME)
}

func pass(prim string, isCode bool) bool {
	p := strings.ToLower(prim)
	return !isCode && (p == consts.PAIR || p == consts.OR)
}

func getCode(prim string) (string, error) {
	code, ok := codes[prim]
	if ok {
		return code, nil
	}

	for template, code := range regCodes {
		if template[0] != prim[0] {
			continue
		}
		re := regexp.MustCompile(template)
		if re.MatchString(prim) {
			return code, nil
		}
	}
	return "00", errors.Errorf("Unknown primitive: %s", prim)
}
