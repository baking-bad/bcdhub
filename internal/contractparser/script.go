package contractparser

import (
	"fmt"
)

// Script -
type Script struct {
	Code    Code
	Storage Storage

	Tags               Set
	HardcodedAddresses Set
}

// New -
func New(script map[string]interface{}) (s Script, err error) {
	code, err := newCode(script)
	if err != nil {
		return
	}
	s.Code = code

	store, ok := script["storage"]
	if !ok {
		return s, fmt.Errorf("Can't find tag 'storage'")
	}
	s.Storage, err = newStorage(store)
	if err != nil {
		return s, fmt.Errorf("newStorage: %v", err)
	}

	hardcoded, err := FindHardcodedAddresses(script)
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
