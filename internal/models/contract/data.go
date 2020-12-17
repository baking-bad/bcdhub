package contract

import "time"

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

// DiffTask -
type DiffTask struct {
	Network1 string
	Address1 string
	Network2 string
	Address2 string
}

// Stats -
type Stats struct {
	TxCount        int64     `json:"tx_count"`
	LastAction     time.Time `json:"last_action"`
	Balance        int64     `json:"balance"`
	TotalWithdrawn int64     `json:"total_withdrawn"`
}

// DAppStats -
type DAppStats struct {
	Users  int64 `json:"users"`
	Calls  int64 `json:"txs"`
	Volume int64 `json:"volume"`
}

// Address -
type Address struct {
	Address string
	Network string
}
