package rawbytes

import (
	"fmt"
	"strings"
)

// ToMicheline -
func ToMicheline(input string) (string, error) {
	str, offset, err := hexToMicheline(input)
	if err != nil {
		return "", err
	}

	if offset != len(input) {
		return "", fmt.Errorf("invalid offset %v, expected %v", offset, len(input))
	}

	return str, err
}

func hexToMicheline(hex string) (string, int, error) {
	var code string
	var offset int

	if len(hex) < 2 {
		return code, offset, fmt.Errorf("hexToMicheline error. input less than 2. input: %v", hex)
	}

	fieldType := hex[:2]
	offset += 2

	switch fieldType {
	case "00":
		val, length, err := decodeInt(hex[offset:])
		if err != nil {
			return code, offset, err
		}
		code += fmt.Sprintf(`{ "int": "%v" }`, val)
		offset += length
	case "01":
		val, length, err := decodeString(hex[offset:])
		if err != nil {
			return code, offset, err
		}
		code += fmt.Sprintf(`{ "string": "%v" }`, val)
		offset += length
	case "02":
		val, length, err := decodeArray(hex[offset:], 0)
		if err != nil {
			return code, offset, err
		}
		code += val
		offset += length
	case "03":
		prim, err := decodePrim(hex[offset : offset+2])
		if err != nil {
			return code, offset, err
		}
		code += fmt.Sprintf(`{ "prim": "%v" }`, prim)
		offset += 2
	case "04":
		prim, err := decodePrim(hex[offset : offset+2])
		if err != nil {
			return code, offset, err
		}
		offset += 2

		annots, length, err := decodeAnnotations(hex[offset:])
		if err != nil {
			return code, offset, err
		}
		code += fmt.Sprintf(`{ "prim": "%v", "annots": [ %v ] }`, prim, annots)
		offset += length
	case "05":
		prim, err := decodePrim(hex[offset : offset+2])
		if err != nil {
			return code, offset, err
		}
		offset += 2

		args, length, err := hexToMicheline(hex[offset:])
		if err != nil {
			return code, offset, err
		}
		code += fmt.Sprintf(`{ "prim": "%v", "args": [ %v ] }`, prim, args)
		offset += length
	case "06":
		prim, err := decodePrim(hex[offset : offset+2])
		if err != nil {
			return code, offset, err
		}
		offset += 2

		args, length, err := hexToMicheline(hex[offset:])
		if err != nil {
			return code, offset, err
		}
		offset += length

		annots, length, err := decodeAnnotations(hex[offset:])
		if err != nil {
			return code, offset, err
		}
		code += fmt.Sprintf(`{ "prim": "%v", "args": [ %v ], "annots": [ %v ] }`, prim, args, annots)
		offset += length
	case "07":
		prim, err := decodePrim(hex[offset : offset+2])
		if err != nil {
			return code, offset, err
		}
		offset += 2

		args := make([]string, 0, 2)

		arg1, length, err := hexToMicheline(hex[offset:])
		if err != nil {
			return code, offset, err
		}
		args = append(args, arg1)
		offset += length

		arg2, length, err := hexToMicheline(hex[offset:])
		if err != nil {
			return code, offset, err
		}
		args = append(args, arg2)
		offset += length

		code += fmt.Sprintf(`{ "prim": "%v", "args": [ %v ] }`, prim, strings.Join(args, ", "))
	case "08":
		prim, err := decodePrim(hex[offset : offset+2])
		if err != nil {
			return code, offset, err
		}
		offset += 2

		args := make([]string, 0, 2)

		arg1, length, err := hexToMicheline(hex[offset:])
		if err != nil {
			return code, offset, err
		}
		args = append(args, arg1)
		offset += length

		arg2, length, err := hexToMicheline(hex[offset:])
		if err != nil {
			return code, offset, err
		}
		args = append(args, arg2)
		offset += length

		annots, length, err := decodeAnnotations(hex[offset:])
		if err != nil {
			return code, offset, err
		}

		code += fmt.Sprintf(`{ "prim": "%v", "args": [ %v ], "annots": [ %v ] }`, prim, strings.Join(args, ", "), annots)
		offset += length
	case "09":
		prim, err := decodePrim(hex[offset : offset+2])
		if err != nil {
			return code, offset, err
		}
		offset += 2

		args, length, err := decodeArray(hex, offset)
		if err != nil {
			return code, offset, err
		}
		offset += length - 4

		if len(hex) < offset+8 {
			return code, offset, fmt.Errorf("hexToMicheline err. Input too short: %v", hex)
		}

		if hex[offset:offset+8] != "00000000" {
			annots, length, err := decodeAnnotations(hex[offset:])
			if err != nil {
				return code, offset, err
			}

			code += fmt.Sprintf(`{ "prim": "%v", "args": %v, "annots": [ %v ] }`, prim, args, annots)
			offset += length
		} else {
			code += fmt.Sprintf(`{ "prim": "%v", "args": %v }`, prim, args)
			offset += 8
		}
	case "0a":
		if len(hex) < offset+8 {
			return code, offset, fmt.Errorf("hexToMicheline err. Input too short: %v", hex)
		}

		val := hex[offset+8:]
		length := len(hex[offset+8:]) + 8
		code += fmt.Sprintf(`{ "bytes": "%v" }`, val)
		offset += length
	default:
		return code, offset, fmt.Errorf("Unknown prefix %v", hex[:2])
	}

	return code, offset, nil
}
