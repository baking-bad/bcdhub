package rawbytes

import (
	"fmt"
	"strings"
)

type arrayDecoder struct{}

// Decode -
func (d arrayDecoder) Decode(dec *decoder, code *strings.Builder) (int, error) {
	code.WriteString("[")
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
