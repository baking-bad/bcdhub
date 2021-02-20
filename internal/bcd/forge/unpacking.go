package forge

import (
	"encoding/hex"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
)

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
	return res, nil
}
