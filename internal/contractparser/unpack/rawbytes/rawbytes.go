package rawbytes

import (
	"fmt"
	"io"
	"strings"
)

var forgers = map[byte]forger{
	0x00: intDecoder{},
	0x01: stringDecoder{},
	0x02: arrayDecoder{},
	0x03: newPrimDecoder(0, false),
	0x04: newPrimDecoder(0, true),
	0x05: newPrimDecoder(1, false),
	0x06: newPrimDecoder(1, true),
	0x07: newPrimDecoder(2, false),
	0x08: newPrimDecoder(2, true),
	0x09: primGeneral{},
	0x0a: bytesDecoder{},
}

// ToMicheline -
func ToMicheline(input string) (string, error) {
	dec := newDecoder(strings.NewReader(input))
	var code strings.Builder
	if _, err := hexToMicheline(dec, &code); err != nil {
		return "", err
	}

	b := make([]byte, 1)
	if _, err := dec.Read(b); err != nil && err != io.EOF {
		return "", err
	} else if err == nil {
		return "", fmt.Errorf("input is not empty")
	}

	return code.String(), nil
}

func hexToMicheline(dec *decoder, code *strings.Builder) (int, error) {
	ft := make([]byte, 1)
	if n, err := dec.Read(ft); err != nil {
		return n, err
	}

	if f, ok := forgers[ft[0]]; ok {
		return f.Decode(dec, code)
	}
	return 1, fmt.Errorf("Unknown type: %x", ft[0])
}
