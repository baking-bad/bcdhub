package rawbytes

import (
	"fmt"
	"io"
	"strings"
)

type bytesDecoder struct{}

// Decode -
func (d bytesDecoder) Decode(dec io.Reader, code *strings.Builder) (int, error) {
	length, err := decodeLength(dec)
	if err != nil {
		return 4, err
	}
	data := make([]byte, length)
	if _, err := dec.Read(data); err != nil && err != io.EOF {
		return 4 + length, err
	}
	fmt.Fprintf(code, `{ "bytes": "%x" }`, data)
	return 4 + length, nil
}
