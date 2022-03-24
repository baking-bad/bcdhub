package contract

import "github.com/baking-bad/bcdhub/internal/models/types"

// SameResponse -
type SameResponse struct {
	Count     int64      `json:"count"`
	Contracts []Contract `json:"contracts"`
}

// Address -
type Address struct {
	Network types.Network
	Address string
}

// Stats -
type Stats struct {
	SameCount int64
}
