package forge

import (
	"encoding/hex"
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

// CollectStrings - returns strings contained in tree. If `tryUnpack` is true, it tries to unpack bytes value.
func CollectStrings(node *base.Node, tryUnpack bool) ([]string, error) {
	res := make([]string, 0)

	switch {
	case node.StringValue != nil && *node.StringValue != "":
		res = append(res, *node.StringValue)
	case tryUnpack && node.BytesValue != nil:
		val := *node.BytesValue
		if s, err := UnforgeAddress(val); err == nil {
			res = append(res, s)
			return res, nil
		}
		if s, err := UnforgeContract(val); err == nil {
			res = append(res, s)
			return res, nil
		}
		if s, err := UnforgeBakerHash(val); err == nil {
			res = append(res, s)
			return res, nil
		}

		data, err := hex.DecodeString(val)
		if err != nil {
			return nil, err
		}
		tree, err := Unpack(data)
		if err != nil {
			b, err := hex.DecodeString(val)
			if err == nil && isPrintableASCII(string(b)) {
				res = append(res, string(b))
			}
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
	return res, nil
}

// DecodeString -
func DecodeString(str string) string {
	if s, err := UnforgeAddress(str); err == nil {
		return s
	}
	if s, err := UnforgeContract(str); err == nil {
		return s
	}
	if s, err := UnforgeBakerHash(str); err == nil {
		return s
	}
	data, err := hex.DecodeString(str)
	if err != nil {
		return str
	}
	if isPrintableASCII(str) {
		return string(data)
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
