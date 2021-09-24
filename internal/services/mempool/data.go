package mempool

import "encoding/json"

// PendingOperations -
type PendingOperations struct {
	Transactions []PendingTransaction `json:"transactions"`
	Originations []PendingOrigination `json:"originations"`
}

// PendingOrigination -
type PendingOrigination struct {
	Balance         string          `json:"balance"`
	Branch          string          `json:"branch"`
	CreatedAt       int64           `json:"created_at"`
	Delegate        string          `json:"delegate"`
	Errors          json.RawMessage `json:"errors"`
	ExpirationLevel int64           `json:"expiration_level"`
	Fee             int64           `json:"fee"`
	GasLimit        int64           `json:"gas_limit"`
	Kind            string          `json:"kind"`
	Level           int64           `json:"level"`
	Signature       string          `json:"signature"`
	Source          string          `json:"source"`
	Status          string          `json:"status"`
	Storage         json.RawMessage `json:"storage"`
	StorageLimit    int64           `json:"storage_limit"`
	UpdatedAt       int64           `json:"updated_at"`
	Network         string          `json:"network"`
	Hash            string          `json:"hash"`
	Counter         int64           `json:"counter"`
	Raw             json.RawMessage `json:"raw"`
	Protocol        string          `json:"protocol"`
}

// PendingTransaction -
type PendingTransaction struct {
	Amount          json.Number     `json:"amount"`
	Branch          string          `json:"branch"`
	CreatedAt       int64           `json:"created_at"`
	Errors          json.RawMessage `json:"errors"`
	ExpirationLevel *int64          `json:"expiration_level,omitempty"`
	Fee             int64           `json:"fee"`
	GasLimit        int64           `json:"gas_limit"`
	Kind            string          `json:"kind"`
	Level           int64           `json:"level"`
	Parameters      json.RawMessage `json:"parameters"`
	Signature       string          `json:"signature"`
	Source          string          `json:"source"`
	Status          string          `json:"status"`
	StorageLimit    int64           `json:"storage_limit"`
	UpdatedAt       int64           `json:"updated_at"`
	Destination     string          `json:"destination"`
	Network         string          `json:"network"`
	Hash            string          `json:"hash"`
	Counter         int64           `json:"counter"`
	Raw             json.RawMessage `json:"raw"`
	Protocol        string          `json:"protocol"`
}
