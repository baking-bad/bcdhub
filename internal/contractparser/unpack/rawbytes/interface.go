package rawbytes

import (
	"strings"
)

type forger interface {
	Decode(*decoder, *strings.Builder) (int, error)
}
