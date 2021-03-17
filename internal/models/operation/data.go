package operation

// DAppStats -
type DAppStats struct {
	Users  int64 `json:"users"`
	Calls  int64 `json:"txs"`
	Volume int64 `json:"volume"`
}
