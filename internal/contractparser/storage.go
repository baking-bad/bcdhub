package contractparser

// Storage -
type Storage struct {
	Value Schema
	Tags  Set
}

func newStorage(storage interface{}) (Storage, error) {
	res := Storage{
		Tags: make(Set),
	}
	res.Value = newSchema(storage)
	return res, nil
}
