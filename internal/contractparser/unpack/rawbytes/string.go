package rawbytes

import (
	"encoding/hex"
	"fmt"
	"strconv"
)

func decodeString(h string) (string, int, error) {
	if len(h) < 8 {
		return "", 0, fmt.Errorf("decodeString err. input too short: %v", h)
	}

	length, err := strconv.ParseInt(h[:8], 16, 64)
	if err != nil {
		return "", 0, err
	}

	offset := 8 + int(length)*2

	if len(h) < offset {
		return "", 0, fmt.Errorf("decodeString err. input too short: %v", h)
	}

	data, err := hex.DecodeString(h[8:offset])
	if err != nil {
		return "", 0, err
	}

	return string(data), offset, nil
}
