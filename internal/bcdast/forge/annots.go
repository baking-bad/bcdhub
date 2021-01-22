package forge

import (
	"strings"
)

type annots struct {
	Value []string
}

// Unforge -
func (a *annots) Unforge(data []byte) (int, error) {
	s := String{}
	n, err := s.Unforge(data)
	if err != nil {
		return n, err
	}

	if s.StringValue != nil && len(*s.StringValue) > 0 {
		a.Value = strings.Split(*s.StringValue, " ")
	}

	return n, nil
}
