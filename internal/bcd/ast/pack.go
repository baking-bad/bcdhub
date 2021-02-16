package ast

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/forge"
)

// Pack -
func Pack(node Base) (string, error) {
	data, err := Forge(node, true)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s", forge.PackPrefix, data), nil
}
