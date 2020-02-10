package rawbytes

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

func decodeArray(hex string, offset int) (string, int, error) {
	var code string

	if len(hex) < offset+8 {
		return code, offset, fmt.Errorf("decodeArray err. input too short: %v", hex)
	}

	length, err := strconv.ParseInt(hex[offset:offset+8], 16, 64)
	if err != nil {
		log.Fatal(err)
	}
	offset += 8

	var buffer []string
	var consumed int

	for consumed < int(length) {
		c, o, err := hexToMicheline(hex[offset:])
		if err != nil {
			return code, offset, err
		}
		buffer = append(buffer, c)
		consumed += o / 2
		offset += o
	}

	if length == 0 {
		code += `[]`
	} else {
		code += fmt.Sprintf(`[ %v ]`, strings.Join(buffer, ", "))
	}

	return code, offset, nil
}
