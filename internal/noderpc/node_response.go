package noderpc

import (
	"time"
)

// Header is a header in a block returned by the Tezos RPC API.
type Header struct {
	Level       int64     `json:"level"`
	Protocol    string    `json:"protocol"`
	Timestamp   time.Time `json:"timestamp"`
	ChainID     string    `json:"chain_id"`
	Hash        string    `json:"hash"`
	Predecessor string    `json:"predecessor"`
}

// Constants -
type Constants struct {
	CostPerByte                  int64   `json:"cost_per_byte"`
	HardGasLimitPerOperation     int64   `json:"hard_gas_limit_per_operation"`
	HardStorageLimitPerOperation int64   `json:"hard_storage_limit_per_operation"`
	TimeBetweenBlocks            []int64 `json:"time_between_blocks"`
}
