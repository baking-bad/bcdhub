package contract

import (
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/parsers"
)

// Parser -
type Parser interface {
	Parse(operation *operation.Operation, store parsers.Store) error
}
