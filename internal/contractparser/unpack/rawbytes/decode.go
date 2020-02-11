package rawbytes

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

func decodeAnnots(dec *decoder) (string, int, error) {
	sb := make([]byte, 4)
	if n, err := dec.Read(sb); err != nil {
		return "", n, err
	}

	var annots string
	length := int(binary.BigEndian.Uint32(sb))
	if length != 0 {
		data := make([]byte, length)
		if n, err := dec.Read(data); err != nil && err != io.EOF {
			return "", n, err
		}

		var ret []string
		for _, v := range strings.Split(string(data), " ") {
			ret = append(ret, v)
		}

		annots = strings.Join(ret, `", "`)
	}
	return annots, length + 4, nil
}

func decodeLength(dec *decoder) (int, error) {
	b := make([]byte, 4)
	if n, err := dec.Read(b); err != nil {
		return n, err
	}

	length := int(binary.BigEndian.Uint32(b))
	return length, nil
}

func decodePrim(dec *decoder) (string, error) {
	b := make([]byte, 1)
	if _, err := dec.Read(b); err != nil {
		return "", err
	}
	key := int(b[0])
	if key > len(primKeywords) {
		return "", &invalidDataError{
			message: fmt.Sprintf("invalid prim keyword %x", b),
			typ:     "prim",
		}
	}
	return primKeywords[key], nil
}

func decodeArgs(dec *decoder, code *strings.Builder, count int) (length int, err error) {
	for i := 0; i < count; i++ {
		n, err := hexToMicheline(dec, code)
		if err != nil {
			return n + 1, err
		}
		length += n + 1
		if count != i+1 {
			code.WriteString(", ")
		}
	}
	return
}
