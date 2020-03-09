package contractparser

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/language"

	"github.com/aopoltorzhicky/bcdhub/internal/helpers"
	"github.com/tidwall/gjson"
)

// Script -
type Script struct {
	Code    Code
	Storage Storage

	Tags               helpers.Set
	HardcodedAddresses helpers.Set
}

// New -
func New(script gjson.Result) (s Script, err error) {
	code, err := newCode(script)
	if err != nil {
		return
	}
	s.Code = code

	s.Storage, err = newStorage(script.Get("storage"))
	if err != nil {
		return s, fmt.Errorf("newStorage: %v", err)
	}

	hardcoded, err := FindHardcodedAddresses(script.Get("code"))
	if err != nil {
		return
	}
	s.HardcodedAddresses = hardcoded
	s.Tags = make(helpers.Set)

	return
}

// Parse -
func (s *Script) Parse() {
	s.getTags()
}

// Language -
func (s *Script) Language() string {
	if s.Code.Language == language.LangUnknown {
		entrypoints, _ := s.Code.Parameter.Metadata.GetEntrypoints()
		return language.DetectInEntries(entrypoints)
	}

	return s.Code.Language
}

func (s *Script) getTags() {
	s.Tags.Append(s.Code.Tags.Values()...)
	s.Tags.Append(s.Storage.Tags.Values()...)
	s.Tags.Append(s.Code.Parameter.Tags.Values()...)
}
