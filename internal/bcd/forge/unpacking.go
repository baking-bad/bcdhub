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

// TryUnpackString - try unpack `str` to tree. If try failed returns `nil`
func TryUnpackString(str string) []*base.Node {
	tree, err := UnpackString(str)
	if err == nil {
		for i := range tree {
			tree[i] = tryUnpackNode(tree[i])
		}
		return tree
	}
	return nil
}

func tryUnpackNode(node *base.Node) *base.Node {
	if node.BytesValue == nil {
		for i := range node.Args {
			node.Args[i] = tryUnpackNode(node.Args[i])
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

	unpacked := TryUnpackString(value)
	if len(unpacked) > 0 {
		node = unpacked[0]
	}
	return node
}

// CollectStrings - returns strings contained in tree. If `tryUnpack` is true, it tries to unpack bytes value.
func CollectStrings(node *base.Node, tryUnpack bool) ([]string, error) {
	res := make([]string, 0)

	switch {
	case node.StringValue != nil && *node.StringValue != "":
		res = append(res, *node.StringValue)
	case tryUnpack && node.BytesValue != nil:
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
			resArg, err := CollectStrings(tree[i], tryUnpack)
			if err != nil {
				return res, nil
			}
			res = append(res, resArg...)
		}
	case len(node.Args) > 0:
		for i := range node.Args {
			argRes, err := CollectStrings(node.Args[i], tryUnpack)
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
