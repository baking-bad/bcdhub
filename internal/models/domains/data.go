package domains

import (
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
)

// BigMapDiff -
type BigMapDiff struct {
	*bigmapdiff.BigMapDiff

	Operation *operation.Operation `pg:"rel:has-one"`
	Protocol  *protocol.Protocol   `pg:"rel:has-one"`
}

// Same -
type Same struct {
	contract.Contract
	Network string
}
