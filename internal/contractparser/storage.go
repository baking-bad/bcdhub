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
		args, ok := t["args"]
		if !ok {
			args = []interface{}{}
		}
		for _, a := range args.([]interface{}) {
			if err := s.parseItem(a); err != nil {
				return err
			}
		}

		s.handlePrimitive(t)
	default:
		return fmt.Errorf("Unknown value type: %T", t)
	}
	return nil
}

func (s *Storage) handlePrimitive(obj map[string]interface{}) (err error) {
	if err = s.detectLanguage(obj); err != nil {
		return
	}
	s.findTags(obj)
	return
}

func (s *Storage) detectLanguage(obj map[string]interface{}) error {
	if s.Language != LangUnknown {
		return nil
	}

	if detectPython(obj) {
		s.Language = LangPython
		return nil
	}
	return nil
}

func (s *Storage) findTags(obj map[string]interface{}) {
	tag := primTags(obj)
	_, ok := s.Tags[tag]
	if tag != "" && !ok {
		s.Tags[tag] = struct{}{}
	}
}
