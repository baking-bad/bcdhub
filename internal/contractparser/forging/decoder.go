package forging

import (
	"encoding/hex"
	"fmt"
	"io"
)

// bufferSize is the number of hexadecimal characters to buffer in encoder and decoder.
const bufferSize = 32768

type decoder struct {
	r   io.Reader
	err error
	in  []byte
	arr [bufferSize]byte
}

type invalidDataError struct {
	message string
	typ     string
}

func (e *invalidDataError) Error() string {
	return fmt.Sprintf("Invalid data: [%s] %s", e.typ, e.message)
}

func newDecoder(r io.Reader) *decoder {
	return &decoder{r: r}
}

func (d *decoder) Read(p []byte) (n int, err error) {
	// Fill internal buffer with sufficient bytes to decode
	if len(d.in) < 2 && d.err == nil {
		var numCopy, numRead int
		numCopy = copy(d.arr[:], d.in) // Copies either 0 or 1 bytes
		numRead, d.err = d.r.Read(d.arr[numCopy:])
		d.in = d.arr[:numCopy+numRead]
		if d.err == io.EOF && len(d.in)%2 != 0 {
			if _, ok := fromHexChar(d.in[len(d.in)-1]); !ok {
				d.err = hex.InvalidByteError(d.in[len(d.in)-1])
			} else {
				d.err = io.ErrUnexpectedEOF
			}
		}
	}

	// Decode internal buffer into output buffer
	if numAvail := len(d.in) / 2; len(p) > numAvail {
		p = p[:numAvail]
	}
	numDec, err := hex.Decode(p, d.in[:len(p)*2])
	d.in = d.in[2*numDec:]
	if err != nil {
		d.in, d.err = nil, err // Decode error; discard input remainder
	}

	if len(d.in) < 2 {
		return numDec, d.err // Only expose errors when buffer fully consumed
	}
	return numDec, nil
}

func (d *decoder) Len() int {
	return len(d.in)
}

// fromHexChar converts a hex character into its value and a success flag.
func fromHexChar(c byte) (byte, bool) {
	switch {
	case '0' <= c && c <= '9':
		return c - '0', true
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10, true
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10, true
	}

	return 0, false
}
