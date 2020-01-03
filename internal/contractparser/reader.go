package contractparser

import "io"

// HashReader -
type HashReader struct {
	buf []byte
}

// NewHashReader -
func NewHashReader() *HashReader {
	return &HashReader{
		buf: make([]byte, 0),
	}
}

// Read -
func (r *HashReader) Read(p []byte) (n int, err error) {
	n = copy(p, r.buf)
	r.buf = (r.buf)[n:]
	if n == 0 {
		return 0, io.EOF
	}
	return n, nil
}

// ReadByte -
func (r *HashReader) ReadByte() (byte, error) {
	if len(r.buf) == 0 {
		return 0, io.EOF
	}
	res := r.buf[0]
	r.buf = (r.buf)[1:]
	return res, nil
}

// WriteString -
func (r *HashReader) WriteString(s string) error {
	b := []byte(s)
	r.buf = append(r.buf, b...)
	return nil
}
