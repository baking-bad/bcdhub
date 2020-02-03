package models

import "time"

// Operation -
type Operation struct {
	ID string `json:"-"`

	Network  string `json:"network"`
	Protocol string `json:"protocol"`
	Hash     string `json:"hash"`
	Internal bool   `json:"internal"`

	Timestamp     time.Time `json:"timestamp"`
	Level         int64     `json:"level"`
	Kind          string    `json:"kind"`
	Source        string    `json:"source"`
	Fee           int64     `json:"fee,omitempty"`
	Counter       int64     `json:"counter,omitempty"`
	GasLimit      int64     `json:"gas_limit,omitempty"`
	StorageLimit  int64     `json:"storage_limit,omitempty"`
	Amount        int64     `json:"amount,omitempty"`
	Destination   string    `json:"destination,omitempty"`
	PublicKey     string    `json:"public_key,omitempty"`
	ManagerPubKey string    `json:"manager_pubkey,omitempty"`
	Balance       int64     `json:"balance,omitempty"`
	Delegate      string    `json:"delegate,omitempty"`
	Parameters    string    `json:"parameters,omitempty"`

	BalanceUpdates []BalanceUpdate  `json:"balance_updates,omitempty"`
	Result         *OperationResult `json:"result,omitempty"`

	DeffatedStorage string `json:"deffated_storage,omitempty"`
}

// BalanceUpdate -
type BalanceUpdate struct {
	Kind     string `json:"kind"`
	Contract string `json:"contract,omitempty"`
	Change   int64  `json:"change"`
	Category string `json:"category,omitempty"`
	Delegate string `json:"delegate,omitempty"`
	Cycle    int    `json:"cycle,omitempty"`
}

// OperationResult -
type OperationResult struct {
	Status              string `json:"status"`
	ConsumedGas         int64  `json:"consumed_gas,omitempty"`
	StorageSize         int64  `json:"storage_size,omitempty"`
	PaidStorageSizeDiff int64  `json:"paid_storage_size_diff,omitempty"`
	Originated          string `json:"-"`
	Errors              string `json:"errors,omitempty"`

	BalanceUpdates []BalanceUpdate `json:"balance_updates,omitempty"`
}
