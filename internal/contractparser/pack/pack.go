package pack

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// Micheline pack micheline-json to bytes
func Micheline(node gjson.Result) ([]byte, error) {
	if node.IsArray() {
		return packArray(node)
	}

	if node.IsObject() {
		return packObject(node)
	}

	return nil, errors.Errorf("invalid micheline data %v", node)
}

func packArray(node gjson.Result) ([]byte, error) {
	var result bytes.Buffer
	result.WriteByte(0x02)

	var temp bytes.Buffer

	for _, item := range node.Array() {
		res, err := Micheline(item)
		if err != nil {
			return nil, err
		}

		temp.Write(res)
	}

	result.Write(packArrayWithLength(temp.Bytes()))

	return result.Bytes(), nil
}

func packArrayWithLength(data []byte) []byte {
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, uint32(len(data)))

	return append(bs, data...)
}

func packObject(node gjson.Result) ([]byte, error) {
	if node.Get("prim").Exists() {
		return packObjectPrim(node)
	}

	if node.Get("bytes").Exists() {
		return packObjectBytes(node)
	}

	if node.Get("int").Exists() {
		return packObjectInt(node), nil
	}

	if node.Get("string").Exists() {
		return packObjectString(node)
	}

	return nil, errors.Errorf("some shit happend %v", node)
}

func packObjectPrim(node gjson.Result) ([]byte, error) {
	var result bytes.Buffer
	argsLen := int(node.Get("args.#").Int())
	annotsLen := int(node.Get("annots.#").Int())

	result.WriteByte(lenTags[argsLen][annotsLen > 0])
	result.WriteByte(primTags[node.Get("prim").String()])

	if argsLen > 0 {
		var args bytes.Buffer
		for _, item := range node.Get("args").Array() {
			arg, err := Micheline(item)
			if err != nil {
				return nil, err
			}
			args.Write(arg)
		}

		if argsLen < 3 {
			result.Write(args.Bytes())
		} else {
			result.Write(packArrayWithLength(args.Bytes()))
		}
	}

	if annotsLen > 0 {
		var temp bytes.Buffer
		for i, item := range node.Get("annots").Array() {
			if i != 0 {
				temp.WriteByte(0x20)
			}

			temp.WriteString(item.String())
		}

		result.Write(packArrayWithLength(temp.Bytes()))
	} else {
		result.Write([]byte{0x00, 0x00, 0x00, 0x00})
	}

	return result.Bytes(), nil
}

func packObjectBytes(node gjson.Result) ([]byte, error) {
	var result bytes.Buffer

	if _, err := result.Write([]byte{0x05, 0x0A}); err != nil {
		return nil, err
	}

	data, err := hex.DecodeString(node.Get("bytes").String())
	if err != nil {
		return nil, err
	}

	if _, err := result.Write(packArrayWithLength(data)); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func packObjectInt(node gjson.Result) []byte {
	result := []byte{0x05, 0x00}

	val := node.Get("int").Float()
	i := int(math.Abs(val))
	b := (i & 0x3F)

	if val < 0 {
		b |= 0xC0
	} else {
		b |= 0x80
	}

	result = append(result, byte(b))

	for i >>= 6; i != 0; i >>= 7 {
		result = append(result, byte((i&0x7F)|0x80))
	}

	result[len(result)-1] &= 0x7F

	return result
}

func packObjectString(node gjson.Result) ([]byte, error) {
	var result bytes.Buffer

	if _, err := result.Write([]byte{0x05, 0x01}); err != nil {
		return nil, err
	}

	data := []byte(node.Get("string").String())

	if _, err := result.Write(packArrayWithLength(data)); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}
