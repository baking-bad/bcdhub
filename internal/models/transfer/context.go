package transfer

import "github.com/baking-bad/bcdhub/internal/models/types"

// GetContext -
type GetContext struct {
	Contracts []string
	Network   types.Network
	AccountID int64

	Start       uint
	End         uint
	SortOrder   string
	LastID      string
	Size        int64
	Offset      int64
	TokenID     *uint64
	OperationID *int64
}
