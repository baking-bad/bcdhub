package forging

import (
	"io"
	"strings"

	"github.com/pkg/errors"
)

var forgers = map[byte]forger{
	0x00: intForger{},
	0x01: stringForger{},
	0x02: arrayForger{},
	0x03: newPrimForger(0, false),
	0x04: newPrimForger(0, true),
	0x05: newPrimForger(1, false),
	0x06: newPrimForger(1, true),
	0x07: newPrimForger(2, false),
	0x08: newPrimForger(2, true),
	0x09: primGeneral{},
	0x0a: bytesForger{},
}

// Unforge -
func Unforge(input string) (string, error) {
	dec := newDecoder(strings.NewReader(input))
	var code strings.Builder
	if _, err := hexToMicheline(dec, &code); err != nil {
		return "", err
	}

	b := make([]byte, 1)
	if _, err := dec.Read(b); err != nil && err != io.EOF {
		return "", err
	} else if err == nil {
		return "", errors.Errorf("input is not empty")
	}

	return code.String(), nil
}

func hexToMicheline(dec *decoder, code *strings.Builder) (int, error) {
	ft := make([]byte, 1)
	if n, err := dec.Read(ft); err != nil {
		return n, err
	}

	if f, ok := forgers[ft[0]]; ok {
		// log.Printf("[hexToMicheline] forger: %T\n", f)
		return f.Unforge(dec, code)
	}
	return 1, errors.Errorf("Unknown type: %x", ft[0])
}
