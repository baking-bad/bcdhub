package noderpc

import stdJSON "encoding/json"

type runCodeRequest struct {
	Script     stdJSON.RawMessage `json:"script"`
	Storage    stdJSON.RawMessage `json:"storage"`
	Input      stdJSON.RawMessage `json:"input"`
	Amount     int64              `json:"amount,string"`
	ChainID    string             `json:"chain_id"`
	Balance    string             `json:"balance,omitempty"`
	Gas        int64              `json:"gas,string,omitempty"`
	Source     string             `json:"source,omitempty"`
	Payer      string             `json:"payer,omitempty"`
	Entrypoint string             `json:"entrypoint,omitempty"`
}

type runOperationRequest struct {
	ChainID   string           `json:"chain_id"`
	Operation runOperationItem `json:"operation"`
}

type runOperationItem struct {
	Branch    string                    `json:"branch"`
	Signature string                    `json:"signature"`
	Contents  []runOperationItemContent `json:"contents"`
}

type runOperationItemContent struct {
	Kind         string             `json:"kind"`
	Fee          int64              `json:"fee,string"`
	Counter      int64              `json:"counter,string"`
	GasLimit     int64              `json:"gas_limit,string"`
	StorageLimit int64              `json:"storage_limit"`
	Source       string             `json:"source"`
	Destination  string             `json:"destination"`
	Amount       int64              `json:"amount,string"`
	Parameters   stdJSON.RawMessage `json:"parameters"`
}
