package models

import (
	"time"
)

// ContractCountStats -
type ContractCountStats struct {
	Total     int64
	SameCount int64
}

// NetworkStats -
type NetworkStats struct {
	ContractsCount       uint64
	UniqueContractsCount uint64
	CallsCount           uint64
	FACount              uint64
}

// ContractStats -
type ContractStats struct {
	Count      int64
	LastAction time.Time
}
