package contractparser

import (
	"fmt"
)

// Script -
type Script struct {
	Code    Code
	Storage Storage

	Tags               Set
	HardcodedAddresses []string
}

// New -
func New(contract map[string]interface{}) (s Script, err error) {
	script, ok := contract["script"]
	if !ok {
		return s, fmt.Errorf("Can`t find tag 'script'")
	}
	m, ok := script.(map[string]interface{})
	if !ok {
		return s, fmt.Errorf("Invalid script type: %T", script)
	}

	code, err := newCode(m)
	if err != nil {
		return
	}
	s.Code = code

	store, ok := m["storage"]
	if !ok {
		return s, fmt.Errorf("Can't find tag 'storage'")
	}
	s.Storage, err = newStorage(store)
	if err != nil {
		return
	}

	hardcoded, err := FindHardcodedAddresses(m)
	if err != nil {
		return
	}
	s.HardcodedAddresses = hardcoded
	s.Tags = make(Set)

	return
}

// Parse -
func (s *Script) Parse() {
	s.getTags()
}

// Language -
func (s *Script) Language() string {
	return s.Code.Language
}

func (s *Script) getTags() {
	s.Tags.Append(s.Code.Tags.Values()...)
	s.Tags.Append(s.Storage.Tags.Values()...)
	s.Tags.Append(s.Code.Parameter.Tags.Values()...)
}
