package operation

import "time"

// ContractStats -
type ContractStats struct {
	TxCount    int64     `json:"tx_count"`
	LastAction time.Time `json:"last_action"`
	Balance    int64     `json:"balance"`
}

// DAppStats -
type DAppStats struct {
	Users  int64 `json:"users"`
	Calls  int64 `json:"txs"`
	Volume int64 `json:"volume"`
}
