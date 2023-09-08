package domains

import (
	"github.com/baking-bad/bcdhub/internal/models/contract"
)

//go:generate mockgen -source=$GOFILE -destination=../mock/domains/mock.go -package=domains -typed
type Repository interface {
	BigMapDiffs(lastID, size int64) ([]BigMapDiff, error)

	Same(network string, c contract.Contract, limit, offset int, availiableNetworks ...string) ([]Same, error)
	SameCount(c contract.Contract, availiableNetworks ...string) (int, error)
}
