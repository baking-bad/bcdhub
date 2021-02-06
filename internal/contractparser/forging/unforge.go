package forging

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
)

func unforgeAnnots(dec *decoder) (string, int, error) {
	sb := make([]byte, 4)
	if n, err := dec.Read(sb); err != nil {
		return "", n, err
	}

	var annots string
	length := int(binary.BigEndian.Uint32(sb))
	if length != 0 {
		if dec.Len() < length {
			return "", 4, errors.Errorf("Invalid annots length got %d has %d", length, dec.Len())
		}
		data := make([]byte, length)
		if n, err := dec.Read(data); err != nil && err != io.EOF {
			return "", n, err
		}
		// log.Printf("[decodeAnnots] data: %x\n", data)

		annots = strings.Join(strings.Split(string(data), " "), `", "`)
	}
	return annots, length + 4, nil
}

func unforgeLength(dec io.Reader) (int, error) {
	b := make([]byte, 4)
	if n, err := dec.Read(b); err != nil {
		return n, err
	}

	length := int(binary.BigEndian.Uint32(b))

	// log.Printf("[decodeLength] %x | length: %v", b, length)
	return length, nil
}

func unforgePrim(dec io.Reader) (string, error) {
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

func unforgeArgs(dec *decoder, code *strings.Builder, count int) (length int, err error) {
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
