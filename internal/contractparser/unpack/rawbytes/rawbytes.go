package rawbytes

import (
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

var forgers = map[byte]forger{
	0x00: intDecoder{},
	0x01: stringDecoder{},
	0x02: arrayDecoder{},
	0x03: primDecoder{},
	0x04: primAnnotsDecoder{},
	0x05: primArgsDecoder{},
	0x0a: bytesDecoder{},
}

// ToMicheline -
func ToMicheline(input string) (string, error) {
	dec := hex.NewDecoder(strings.NewReader(input))
	var code strings.Builder
	if _, err := hexToMicheline(dec, &code); err != nil {
		return "", err
	}

	return code.String(), nil
}

func hexToMicheline(dec io.Reader, code *strings.Builder) (int, error) {
	ft := make([]byte, 1)
	if n, err := dec.Read(ft); err != nil {
		return n, err
	}

	if f, ok := forgers[ft[0]]; ok {
		return f.Decode(dec, code)
	}
	return 1, fmt.Errorf("Unknown type: %x", ft[0])

	// case "06":
	// 	prim, err := decodePrim(hex[offset : offset+2])
	// 	if err != nil {
	// 		return "", offset, err
	// 	}
	// 	offset += 2

	// 	args, length, err := hexToMicheline(hex[offset:])
	// 	if err != nil {
	// 		return "", offset, err
	// 	}
	// 	offset += length

	// 	annots, length, err := decodeAnnotations(hex[offset:])
	// 	if err != nil {
	// 		return "", offset, err
	// 	}
	// 	fmt.Fprintf(&code, `{ "prim": "%v", "args": [ %v ], "annots": [ %v ] }`, prim, args, annots)
	// 	offset += length
	// case "07":
	// 	prim, err := decodePrim(hex[offset : offset+2])
	// 	if err != nil {
	// 		return "", offset, err
	// 	}
	// 	offset += 2

	// 	args := make([]string, 0, 2)

	// 	arg1, length, err := hexToMicheline(hex[offset:])
	// 	if err != nil {
	// 		return "", offset, err
	// 	}
	// 	args = append(args, arg1)
	// 	offset += length

	// 	arg2, length, err := hexToMicheline(hex[offset:])
	// 	if err != nil {
	// 		return "", offset, err
	// 	}
	// 	args = append(args, arg2)
	// 	offset += length

	// 	fmt.Fprintf(&code, `{ "prim": "%v", "args": [ %v ] }`, prim, strings.Join(args, ", "))
	// case "08":
	// 	prim, err := decodePrim(hex[offset : offset+2])
	// 	if err != nil {
	// 		return "", offset, err
	// 	}
	// 	offset += 2

	// 	args := make([]string, 0, 2)

	// 	arg1, length, err := hexToMicheline(hex[offset:])
	// 	if err != nil {
	// 		return "", offset, err
	// 	}
	// 	args = append(args, arg1)
	// 	offset += length

	// 	arg2, length, err := hexToMicheline(hex[offset:])
	// 	if err != nil {
	// 		return "", offset, err
	// 	}
	// 	args = append(args, arg2)
	// 	offset += length

	// 	annots, length, err := decodeAnnotations(hex[offset:])
	// 	if err != nil {
	// 		return "", offset, err
	// 	}

	// 	fmt.Fprintf(&code, `{ "prim": "%v", "args": [ %v ], "annots": [ %v ] }`, prim, strings.Join(args, ", "), annots)
	// 	offset += length
	// case "09":
	// 	prim, err := decodePrim(hex[offset : offset+2])
	// 	if err != nil {
	// 		return "", offset, err
	// 	}
	// 	offset += 2

	// 	args, length, err := decodeArray(hex, offset)
	// 	if err != nil {
	// 		return "", offset, err
	// 	}
	// 	offset += length - 4

	// 	if len(hex) < offset+8 {
	// 		return "", offset, fmt.Errorf("hexToMicheline err. Input too short: %v", hex)
	// 	}

	// 	if hex[offset:offset+8] != "00000000" {
	// 		annots, length, err := decodeAnnotations(hex[offset:])
	// 		if err != nil {
	// 			return "", offset, err
	// 		}

	// 		fmt.Fprintf(&code, `{ "prim": "%v", "args": %v, "annots": [ %v ] }`, prim, args, annots)
	// 		offset += length
	// 	} else {
	// 		fmt.Fprintf(&code, `{ "prim": "%v", "args": %v }`, prim, args)
	// 		offset += 8
	// 	}
}
