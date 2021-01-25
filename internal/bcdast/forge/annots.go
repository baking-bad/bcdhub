package forge

import (
	"strings"
)

type annots struct {
	Value []string
}

func newAnnots() *annots {
	return &annots{
		Value: make([]string, 0),
	}
}

// Unforge -
func (a *annots) Unforge(data []byte) (int, error) {
	s := new(String)
	n, err := s.Unforge(data)
	if err != nil {
		return n, err
	}

	if s.StringValue != nil && len(*s.StringValue) > 0 {
		a.Value = strings.Split(*s.StringValue, " ")
	}

	return n, nil
}

// Forge -
func (a *annots) Forge() ([]byte, error) {
	val := strings.Join(a.Value, " ")
	data := []byte(val)
	l := new(length)
	l.Value = len(data)
	lData, err := l.Forge()
	if err != nil {
		return nil, err
	}
	return append(lData, data...), nil
}
