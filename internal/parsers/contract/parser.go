package contract

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/parsers"
)

// Parser -
type Parser interface {
	Parse(ctx context.Context, operation *operation.Operation, store parsers.Store) error
}
