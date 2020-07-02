package contractparser

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/language"
	"github.com/baking-bad/bcdhub/internal/contractparser/macros"
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
func (s *Script) Language() (string, error) {
	result := language.LangUnknown

	possibleLanguages := s.Code.Language
	possibleLanguages.Merge(s.Code.Parameter.Language)
	possibleLanguages.Merge(s.Code.Storage.Language)

	for _, lang := range language.Priorities() {
		if _, ok := possibleLanguages[lang]; ok {
			result = lang
			break
		}
	}

	if result == language.LangSCaml || result == language.LangUnknown {
		hasMacros, err := macros.HasMacros(s.Code.Code, macros.GetLangugageFamilies())
		if err != nil {
			return result, err
		}

		if hasMacros {
			result = language.LangMichelson
		}
	}

	return result, nil
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
