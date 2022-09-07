package domains

import (
	"github.com/baking-bad/bcdhub/internal/models/contract"
)

// Repository -
type Repository interface {
	BigMapDiffs(lastID, size int64) ([]BigMapDiff, error)

	Same(network string, c contract.Contract, limit, offset int, availiableNetworks ...string) ([]Same, error)
	SameCount(c contract.Contract, availiableNetworks ...string) (int, error)
}
