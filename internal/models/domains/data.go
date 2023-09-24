package domains

import (
	"github.com/baking-bad/bcdhub/internal/models/contract"
)

// Same -
type Same struct {
	contract.Contract
	Network string
}
