package handlers

import "github.com/aopoltorzhicky/bcdhub/internal/models"

// Operation -
type Operation struct {
	ID       string `json:"-"`
	Protocol string `json:"protocol"`
	Hash     string `json:"hash"`
	Internal bool   `json:"internal"`
	Network  string `json:"network"`

	Level         int64  `json:"level"`
	Kind          string `json:"kind"`
	Source        string `json:"source"`
	Fee           int64  `json:"fee,omitempty"`
	Counter       int64  `json:"counter,omitempty"`
	GasLimit      int64  `json:"gas_limit,omitempty"`
	StorageLimit  int64  `json:"storage_limit,omitempty"`
	Amount        int64  `json:"amount,omitempty"`
	Destination   string `json:"destination,omitempty"`
	PublicKey     string `json:"public_key,omitempty"`
	ManagerPubKey string `json:"manager_pubkey,omitempty"`
	Balance       int64  `json:"balance,omitempty"`
	Delegate      string `json:"delegate,omitempty"`

	BalanceUpdates []models.BalanceUpdate  `json:"balance_updates,omitempty"`
	Result         *models.OperationResult `json:"result,omitempty"`

	Parameters  interface{} `json:"parameters,omitempty"`
	StorageDiff interface{} `json:"storage_diff,omitempty"`
}
