package decode

import (
	"encoding/hex"
	"strconv"
)

func decodeString(h string) (string, int, error) {
	length, err := strconv.ParseInt(h[:8], 16, 64)
	if err != nil {
		return "", 0, err
	}

	offset := 8 + int(length)*2

	data, err := hex.DecodeString(h[8:offset])
	if err != nil {
		return "", 0, err
	}

	return string(data), offset, nil
}
