package decode

import (
	"fmt"
	"strings"
)

func decodeAnnotations(h string) (string, int, error) {
	s, offset, err := decodeString(h)
	if err != nil {
		return "", 0, err
	}

	var ret []string

	for _, v := range strings.Split(s, " ") {
		ret = append(ret, fmt.Sprintf(`"%v"`, v))
	}

	return strings.Join(ret, ", "), offset, nil
}
