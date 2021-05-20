package tokenmetadata

import "github.com/baking-bad/bcdhub/internal/models/types"

// GetContext -
type GetContext struct {
	Contract string
	Network  types.Network
	TokenID  *uint64
	MaxLevel int64
	MinLevel int64
	Creator  string
	Name     string
}
