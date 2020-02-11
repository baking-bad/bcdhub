package rawbytes

import (
	"encoding/binary"
	"io"
	"strings"
)

type arrayDecoder struct{}

// Decode -
func (d arrayDecoder) Decode(dec io.Reader, code *strings.Builder) (int, error) {
	code.WriteString("[ ")
	b := make([]byte, 4)
	if n, err := dec.Read(b); err != nil {
		return n, err
	}

	length := int(binary.BigEndian.Uint32(b))
	if length != 0 {
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
	}

	code.WriteString(" ]")
	return length + 4, nil
}
