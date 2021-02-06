package ast

import "github.com/baking-bad/bcdhub/internal/bcd/consts"

// Storage -
type Storage struct {
	*SectionType
}

// NewStorage -
func NewStorage(depth int) *Storage {
	return &Storage{
		SectionType: NewSectionType(consts.STORAGE, depth),
	}
}
