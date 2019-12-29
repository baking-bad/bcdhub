package contractparser

import (
	"encoding/json"
)

// Script -
type Script struct {
	Code    Code
	Storage Storage

	Tags               map[string]struct{}
	HardcodedAddresses []string
}

// New -
func New(script []byte, labels map[string]int) (s Script, err error) {
	var m map[string]interface{}
	if err = json.Unmarshal(script, &m); err != nil {
		return
	}

	code, err := newCode(m, labels)
	if err != nil {
		return
	}
	s.Code = code

	storage, err := newStorage(m)
	if err != nil {
		return
	}
	s.Storage = storage
	s.HardcodedAddresses = findHardcodedAddresses(string(script))

	return
}

// Print -
func (s *Script) Print() {
	s.Code.print()
}

// Entrypoints - returns script entrypoints
func (s *Script) Entrypoints() []Entrypoint {
	return s.Code.entrypoints()
}

// Parse -
func (s *Script) Parse() error {
	if err := s.Code.parseCodePart(); err != nil {
		return err
	}
	if err := s.Storage.parse(); err != nil {
		return err
	}
	s.getTags()

	return nil
}

// Language -
func (s *Script) Language() string {
	if langPriorities[s.Code.Language] > langPriorities[s.Storage.Language] {
		return s.Code.Language
	}
	return s.Storage.Language
}

// Kind - return script kind
func (s *Script) Kind() string {
	switch s.Code.HashCode {
	case HashTestContract:
		return KindTest
	case HashDelegatorContract:
		return KindDelegator
	case HashVestedContract:
		return KindVested
	default:
		return KindSmart
	}
}

func (s *Script) getTags() {
	s.Tags = s.Code.Tags
	for k := range s.Storage.Tags {
		if _, ok := s.Tags[k]; !ok {
			s.Tags[k] = struct{}{}
		}
	}
	for _, k := range endpointsTags(s.Entrypoints()) {
		if _, ok := s.Tags[k]; !ok {
			s.Tags[k] = struct{}{}
		}
	}
}
