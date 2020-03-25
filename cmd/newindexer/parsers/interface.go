package parsers

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

// Parser -
type Parser interface {
	Parse(gjson.Result) (models.Operation, error)
}
