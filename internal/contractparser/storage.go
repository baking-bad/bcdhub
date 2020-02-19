package contractparser

import (
	"github.com/aopoltorzhicky/bcdhub/internal/helpers"
	"github.com/tidwall/gjson"
)

// Storage -
type Storage struct {
	*parser

	Tags helpers.Set
}

func newStorage(storage gjson.Result) (Storage, error) {
	res := Storage{
		parser: &parser{},
		Tags:   make(helpers.Set),
	}
	err := res.parse(storage)
	return res, err
}
