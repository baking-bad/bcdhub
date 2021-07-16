package bigmap

import "github.com/baking-bad/bcdhub/internal/models/types"

// Bucket -
type Bucket struct {
	Diff

	KeysCount int64
}

// Stats -
type Stats struct {
	Total  int64
	Active int64
}

// GetContext -
type GetContext struct {
	Network      types.Network
	Ptr          *int64
	Query        string
	Size         int64
	Offset       int64
	MaxLevel     *int64
	MinLevel     *int64
	CurrentLevel *int64
	Contract     string
}
