package rawbytes

import (
	"io"
	"strings"
)

type arrayDecoder struct{}

// Decode -
func (d arrayDecoder) Decode(dec io.Reader, code *strings.Builder) (int, error) {
	code.WriteString("[")
	length, err := decodeLength(dec)
	if err != nil {
		return 4, err
	}
	if length != 0 {
		code.WriteString(" ")
		var count int
		for count < length {
			n, err := hexToMicheline(dec, code)
			if err != nil {
				return length + 4, err
			}
			count += n + 1
			if count < length {
				code.WriteString(", ")
			}
		}
		code.WriteString(" ")
	}

	code.WriteString("]")
	return length + 4, nil
}
