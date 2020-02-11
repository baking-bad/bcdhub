package rawbytes

import (
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

type bytesDecoder struct{}

// Decode -
func (d bytesDecoder) Decode(dec *decoder, code *strings.Builder) (int, error) {
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

	s := hex.EncodeToString(data)
	intDec := newDecoder(strings.NewReader(s))

	var bufBuilder strings.Builder
	l, err := hexToMicheline(intDec, &bufBuilder)
	if err != nil || intDec.Len() > 0 {
		if _, ok := err.(*invalidDataError); ok || intDec.Len() > 0 {
			fmt.Fprintf(code, `{ "bytes": "%x" }`, data)
			return 4 + length, nil
		}
		return l + 4, err
	}
	fmt.Fprintf(code, `{ "bytes": "%s" }`, bufBuilder.String())
	return 4 + length, nil
}
