package forging

import (
	"strings"
)

type forger interface {
	Unforge(*decoder, *strings.Builder) (int, error)
}
