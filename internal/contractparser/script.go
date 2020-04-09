package contractparser

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/language"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/tidwall/gjson"
)

// Script -
type Script struct {
	Code    Code
	Storage Storage

	Tags               helpers.Set
	Annotations        helpers.Set
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
	s.getAnnotations()
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
	s.Tags = s.Code.Tags
	s.Tags.Merge(s.Storage.Tags)
	s.Tags.Merge(s.Code.Parameter.Tags)
}

func (s *Script) getAnnotations() {
	s.Annotations = s.Code.Annotations
	s.Annotations.Merge(s.Code.Storage.Annotations)
	s.Annotations.Merge(s.Code.Parameter.Annotations)
}
