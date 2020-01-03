package contractparser

import "fmt"

// Storage -
type Storage struct {
	Language string
	Value    interface{}
	Tags     map[string]struct{}
}

func newStorage(script map[string]interface{}) (Storage, error) {
	res := Storage{
		Language: LangUnknown,
		Tags:     make(map[string]struct{}),
	}
	storage, ok := script["storage"]
	if !ok {
		return res, fmt.Errorf("Can't find tag 'storage'")
	}
	res.Value = storage
	return res, nil
}

func (s *Storage) parse() error {
	return s.parseItem(s.Value)
}

func (s *Storage) parseItem(value interface{}) error {
	switch t := value.(type) {
	case []interface{}:
		for _, a := range t {
			if err := s.parseItem(a); err != nil {
				return err
			}
		}
	case map[string]interface{}:
		n := newNode(t)
		for i := range n.Args {
			if err := s.parseItem(n.Args[i]); err != nil {
				return err
			}
		}

		s.handlePrimitive(n)
	default:
		return fmt.Errorf("Unknown value type: %T", t)
	}
	return nil
}

func (s *Storage) handlePrimitive(node *Node) (err error) {
	if err = s.detectLanguage(node); err != nil {
		return
	}
	s.findTags(node)
	return
}

func (s *Storage) detectLanguage(node *Node) error {
	if s.Language != LangUnknown {
		return nil
	}

	if detectPython(node) {
		s.Language = LangPython
		return nil
	}

	if s.Language == "" {
		s.Language = LangUnknown
	}
	return nil
}

func (s *Storage) findTags(node *Node) {
	tag := primTags(node)
	_, ok := s.Tags[tag]
	if tag != "" && !ok {
		s.Tags[tag] = struct{}{}
	}
}
