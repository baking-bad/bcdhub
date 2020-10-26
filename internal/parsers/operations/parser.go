package operations

import (
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/tidwall/gjson"
)

// Parser -
type Parser interface {
	Parse(data gjson.Result) ([]elastic.Model, error)
}
