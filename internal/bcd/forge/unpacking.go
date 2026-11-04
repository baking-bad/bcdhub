package forge

import (
	"encoding/hex"
	"strings"
	"unicode"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Unpack -
func Unpack(data []byte) ([]*base.Node, error) {
	trimmed, err := TrimPackByte(data)
	if err != nil {
		return nil, err
	}
	unforger := NewMichelson()
	_, err = unforger.Unforge(trimmed)
	return unforger.Nodes, err
}

// UnpackString -
func UnpackString(str string) ([]*base.Node, error) {
	data, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}
	return Unpack(data)
}

// MaxUnpackDepth - how many levels of packed bytes nested in packed bytes are
// unpacked automatically. Every level re-decodes and re-encodes nearly the
// whole remaining payload, so unbounded nesting is quadratic in its size.
// Deeper bytes are kept as bytes.
const MaxUnpackDepth = 8

// TryUnpackString - try unpack `str` to tree. If try failed returns `nil`
func TryUnpackString(str string) []*base.Node {
	return tryUnpackString(str, 1)
}

func tryUnpackString(str string, depth int) []*base.Node {
	tree, err := UnpackString(str)
	if err == nil {
		for i := range tree {
			tree[i] = tryUnpackNode(tree[i], depth)
		}
		return tree
	}
	return nil
}

func tryUnpackNode(node *base.Node, depth int) *base.Node {
	if node == nil {
		return nil
	}
	if node.BytesValue == nil {
		for i := range node.Args {
			node.Args[i] = tryUnpackNode(node.Args[i], depth)
		}
		return node
	}

	value := *node.BytesValue
	decoded := tryDecode(value)
	if decoded != "" {
		node.StringValue = &decoded
		node.BytesValue = nil
		return node
	}

	if depth >= MaxUnpackDepth {
		return node
	}
	unpacked := tryUnpackString(value, depth+1)
	if len(unpacked) > 0 {
		node = unpacked[0]
	}
	return node
}

// CollectStrings - returns strings contained in tree. If `tryUnpack` is true, it tries to unpack bytes value.
func CollectStrings(node *base.Node, tryUnpack bool) ([]string, error) {
	return collectStrings(node, tryUnpack, 0)
}

func collectStrings(node *base.Node, tryUnpack bool, depth int) ([]string, error) {
	res := make([]string, 0)
	if node == nil {
		return res, nil
	}

	switch {
	case node.StringValue != nil && *node.StringValue != "":
		res = append(res, *node.StringValue)
	case tryUnpack && node.BytesValue != nil && depth < MaxUnpackDepth:
		val := *node.BytesValue
		decoded := tryDecode(val)
		if decoded != "" {
			res = append(res, decoded)
			return res, nil
		}

		data, err := hex.DecodeString(val)
		if err != nil {
			return nil, err
		}
		tree, err := Unpack(data)
		if err != nil {
			return res, nil
		}
		for i := range tree {
			resArg, err := collectStrings(tree[i], tryUnpack, depth+1)
			if err != nil {
				return res, nil
			}
			res = append(res, resArg...)
		}
	case len(node.Args) > 0:
		for i := range node.Args {
			argRes, err := collectStrings(node.Args[i], tryUnpack, depth)
			if err != nil {
				return nil, err
			}
			res = append(res, argRes...)
		}
	}

	for i := range res {
		res[i] = strings.ReplaceAll(res[i], "\u0000", "\\u0000")
	}
	return res, nil
}

func tryDecode(val string) string {
	buf := strings.TrimPrefix(val, "0x")
	if s, err := UnforgeAddress(buf); err == nil {
		return s
	}
	if s, err := UnforgeContract(buf); err == nil {
		return s
	}
	if s, err := UnforgeBakerHash(buf); err == nil {
		return s
	}

	b, err := hex.DecodeString(val)
	if err == nil && isPrintableASCII(string(b)) {
		return string(b)
	}
	return ""
}

// DecodeString -
func DecodeString(str string) string {
	if s := tryDecode(str); s != "" {
		return s
	}
	data, err := hex.DecodeString(str)
	if err != nil {
		return str
	}
	tree, err := Unpack(data)
	if err != nil {
		return str
	}
	b, err := json.Marshal(tree)
	if err != nil {
		return str
	}
	s, err := formatter.MichelineToMichelsonInline(string(b))
	if err != nil {
		return str
	}
	return s
}

func isPrintableASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII || !unicode.IsPrint(rune(s[i])) {
			return false
		}
	}
	return true
}
