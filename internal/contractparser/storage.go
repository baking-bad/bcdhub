package contractparser

import (
	"github.com/aopoltorzhicky/bcdhub/internal/tlsh"
	"github.com/tidwall/gjson"
)

// Storage -
type Storage struct {
	*parser

	Tags Set
	Hash string
	hash []byte
}

func newStorage(storage gjson.Result) (Storage, error) {
	res := Storage{
		parser: &parser{},
		Tags:   make(Set),
		hash:   make([]byte, 0),
	}
	res.primHandler = res.handlePrimitive
	if err := res.parse(storage); err != nil {
		return res, err
	}

	if len(res.hash) == 0 {
		res.hash = append(res.hash, 0)
	}
	h, err := tlsh.HashBytes(res.hash)
	if err != nil {
		return res, err
	}
	res.Hash = h.String()
	return res, err
}

func (s *Storage) handlePrimitive(n Node) error {
	s.hash = append(s.hash, []byte(n.Prim)...)

	return nil
}
