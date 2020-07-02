package contractparser

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/language"
	"github.com/baking-bad/bcdhub/internal/contractparser/node"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/tidwall/gjson"
)

// Storage -
type Storage struct {
	*parser

	Tags        helpers.Set
	Annotations helpers.Set
	Language    helpers.Set
}

func newStorage(storage gjson.Result) (Storage, error) {
	s := Storage{
		parser:      &parser{},
		Tags:        make(helpers.Set),
		Annotations: make(helpers.Set),
		Language:    make(helpers.Set),
	}

	s.primHandler = s.handlePrimitive
	err := s.parse(storage)

	lang := language.DetectMichelsonInStorage(storage)
	s.Language.Add(lang)
	return s, err
}

func (s *Storage) handlePrimitive(n node.Node) error {
	if n.HasAnnots() {
		s.Annotations.Append(filterAnnotations(n.Annotations)...)
	}
	return nil
}
