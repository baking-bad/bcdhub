package contract

import "github.com/baking-bad/bcdhub/internal/models/types"

// SameResponse -
type SameResponse struct {
	Count     int64      `json:"count"`
	Contracts []Contract `json:"contracts"`
}

// Similar -
type Similar struct {
	*Contract
	Count int64 `json:"count"`
}

// Address -
type Address struct {
	Network types.Network
	Address string
}

// Stats -
type Stats struct {
	SameCount    int64
	SimilarCount int64
}
