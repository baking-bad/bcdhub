package forging

import (
	"fmt"
	"io"
	"strings"
)

type stringForger struct{}

// Decode -
func (d stringForger) Unforge(dec *decoder, code *strings.Builder) (int, error) {
	length, err := unforgeLength(dec)
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
	// log.Printf("[string Decode] data: %x, value: %v\n", data, string(data))
	fmt.Fprintf(code, `{ "string": "%s" }`, data)
	return 4 + length, nil
}
