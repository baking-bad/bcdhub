package rawbytes

import (
	"fmt"
	"io"
	"strings"
)

type stringDecoder struct{}

// Decode -
func (d stringDecoder) Decode(dec io.Reader, code *strings.Builder) (int, error) {
	length, err := decodeLength(dec)
	if err != nil {
		return 4, err
	}
	data := make([]byte, length)
	if _, err := dec.Read(data); err != nil && err != io.EOF {
		return 4 + length, err
	}
	fmt.Fprintf(code, `{ "string": "%s" }`, data)
	return 4 + length, nil
}
