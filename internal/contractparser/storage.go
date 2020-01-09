package contractparser

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/tlsh"
)

// Storage -
type Storage struct {
	fuzzyReader *HashReader

	Tags Set
	Hash string
}

func newStorage(storage interface{}) (Storage, error) {
	res := Storage{
		Tags:        make(Set),
		fuzzyReader: NewHashReader(),
	}
	if err := res.parse(storage); err != nil {
		return res, err
	}

	h, err := tlsh.HashReader(res.fuzzyReader)
	if err != nil {
		return res, err
	}
	res.Hash = h.String()
	return res, err
}

func (s *Storage) parse(v interface{}) error {
	switch t := v.(type) {
	case []interface{}:
		for _, a := range t {
			if err := s.parse(a); err != nil {
				return err
			}
		}
	case map[string]interface{}:
		node := newNode(t)
		for i := range node.Args {
			s.parse(node.Args[i])
		}
		if err := s.handlePrimitive(node); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unknown value type: %T", t)
	}
	return nil
}

func (s *Storage) handlePrimitive(n Node) error {
	if n.Prim != "" {
		s.fuzzyReader.WriteString(n.Prim)
	}
	return nil
}
