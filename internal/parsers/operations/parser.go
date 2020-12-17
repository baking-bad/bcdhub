package operations

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

// Parser -
type Parser interface {
	Parse(data gjson.Result) ([]models.Model, error)
}
