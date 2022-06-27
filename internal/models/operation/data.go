package operation

import "time"

// DAppStats -
type DAppStats struct {
	Users  int64 `json:"users"`
	Calls  int64 `json:"txs"`
	Volume int64 `json:"volume"`
}

// OPG -
type OPG struct {
	LastID       int64     `json:"last_id"`
	ContentIndex int64     `json:"content_index"`
	Counter      int64     `json:"counter"`
	Level        int64     `json:"level"`
	TotalCost    int64     `json:"total_cost"`
	Flow         int64     `json:"flow"`
	Internals    int       `json:"internals"`
	Hash         string    `json:"hash"`
	Entrypoint   string    `json:"entrypoint"`
	Timestamp    time.Time `json:"timestamp"`
}
