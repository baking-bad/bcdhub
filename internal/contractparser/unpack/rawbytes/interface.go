package rawbytes

import (
	"io"
	"strings"
)

type forger interface {
	Decode(io.Reader, *strings.Builder) (int, error)
}
