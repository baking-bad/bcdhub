package rawbytes

import (
	"fmt"
	"io"
	"strings"
)

type stringDecoder struct{}

// Decode -
func (d stringDecoder) Decode(dec *decoder, code *strings.Builder) (int, error) {
	length, err := decodeLength(dec)
	if err != nil {
		return 4, err
	}

	if dec.Len() < length {
		return 4, &invalidDataError{
			typ:     "string",
			message: fmt.Sprintf("Not enough data in reader: %d < %d", dec.Len(), length),
		}
	}
	data := make([]byte, length)
	if _, err := dec.Read(data); err != nil && err != io.EOF {
		return 4 + length, err
	}
	fmt.Fprintf(code, `{ "string": "%s" }`, data)
	return 4 + length, nil
}
