package rawbytes

import (
	"fmt"
	"strings"
)

// HexToMicheline -
func HexToMicheline(hex string) (string, int, error) {
	var code string
	var offset int

	fieldType := hex[offset : offset+2]
	offset += 2

	switch fieldType {
	case "00":
		val, length, err := decodeInt(hex[offset:], 0, true)
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
		val, length, err := decodeArray(hex, offset)
		if err != nil {
			return code, offset, err
		}
		code += val
		offset += length
	case "03":
		prim, err := decodePrim(hex, offset)
		if err != nil {
			return code, offset, err
		}
		code += fmt.Sprintf(`{ "prim": "%v" }`, prim)
		offset += 2
	case "04":
		prim, err := decodePrim(hex, offset)
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
		prim, err := decodePrim(hex, offset)
		if err != nil {
			return code, offset, err
		}
		offset += 2

		args, length, err := HexToMicheline(hex[offset:])
		if err != nil {
			return code, offset, err
		}
		code += fmt.Sprintf(`{ "prim": "%v", "args": [ %v ] }`, prim, args)
		offset += length
	case "06":
		prim, err := decodePrim(hex, offset)
		if err != nil {
			return code, offset, err
		}
		offset += 2

		args, length, err := HexToMicheline(hex[offset:])
		if err != nil {
			return code, offset, err
		}
		offset += length

		annots, anLength, err := decodeAnnotations(hex[offset:])
		if err != nil {
			return code, offset, err
		}
		code += fmt.Sprintf(`{ "prim": "%v", "args": [ %v ], "annots": [ %v ] }`, prim, args, annots)
		offset += anLength
	case "07":
		prim, err := decodePrim(hex, offset)
		if err != nil {
			return code, offset, err
		}
		offset += 2

		args := make([]string, 0, 2)

		arg1, length, err := HexToMicheline(hex[offset:])
		if err != nil {
			return code, offset, err
		}
		args = append(args, arg1)
		offset += length

		arg2, length, err := HexToMicheline(hex[offset:])
		if err != nil {
			return code, offset, err
		}
		args = append(args, arg2)
		offset += length

		code += fmt.Sprintf(`{ "prim": "%v", "args": [ %v ] }`, prim, strings.Join(args, ", "))
		offset += offset
	case "08":
		prim, err := decodePrim(hex, offset)
		if err != nil {
			return code, offset, err
		}
		offset += 2

		args := make([]string, 0, 2)

		arg1, length, err := HexToMicheline(hex[offset:])
		if err != nil {
			return code, offset, err
		}
		args = append(args, arg1)
		offset += length

		arg2, length, err := HexToMicheline(hex[offset:])
		if err != nil {
			return code, offset, err
		}
		args = append(args, arg2)
		offset += length

		annots, anLength, err := decodeAnnotations(hex[offset:])
		if err != nil {
			return code, offset, err
		}

		code += fmt.Sprintf(`{ "prim": "%v", "args": [ %v ], "annots": [ %v ] }`, prim, strings.Join(args, ", "), annots)
		offset += anLength
	case "09":
		prim, err := decodePrim(hex, offset)
		if err != nil {
			return code, offset, err
		}
		offset += 2

		args, length, err := decodeArray(hex, offset)
		if err != nil {
			return code, offset, err
		}
		offset += length - 4

		if hex[offset:offset+8] != "00000000" {
			annots, anLength, err := decodeAnnotations(hex[offset:])
			if err != nil {
				return code, offset, err
			}

			code += fmt.Sprintf(`{ "prim": "%v", "args": %v, "annots": [ %v ] }`, prim, args, annots)
			offset += anLength
		} else {
			code += fmt.Sprintf(`{ "prim": "%v", "args": %v }`, prim, args)
			offset += 8
		}
	case "0a":
		val := hex[offset+8:]
		length := len(hex[offset+8:])*2 + 8
		code += fmt.Sprintf(`{ "bytes": "%v" }`, val)
		offset += length
	default:
		return code, offset, fmt.Errorf("Unknown prefix %v", hex[:2])
	}

	return code, offset, nil
}
