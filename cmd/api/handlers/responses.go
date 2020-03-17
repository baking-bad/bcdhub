package handlers

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/models"
)

// Operation -
type Operation struct {
	ID        string    `json:"-"`
	Protocol  string    `json:"protocol"`
	Hash      string    `json:"hash"`
	Internal  bool      `json:"internal"`
	Network   string    `json:"network"`
	Timesatmp time.Time `json:"timestamp"`

	Level         int64           `json:"level"`
	Kind          string          `json:"kind"`
	Source        string          `json:"source"`
	Fee           int64           `json:"fee,omitempty"`
	Counter       int64           `json:"counter,omitempty"`
	GasLimit      int64           `json:"gas_limit,omitempty"`
	StorageLimit  int64           `json:"storage_limit,omitempty"`
	Amount        int64           `json:"amount,omitempty"`
	Destination   string          `json:"destination,omitempty"`
	PublicKey     string          `json:"public_key,omitempty"`
	ManagerPubKey string          `json:"manager_pubkey,omitempty"`
	Balance       int64           `json:"balance,omitempty"`
	Delegate      string          `json:"delegate,omitempty"`
	Status        string          `json:"status"`
	Entrypoint    string          `json:"entrypoint,omitempty"`
	Errors        []cerrors.Error `json:"errors,omitempty"`

	BalanceUpdates []models.BalanceUpdate  `json:"balance_updates,omitempty"`
	Result         *models.OperationResult `json:"result,omitempty"`

	Parameters  interface{} `json:"parameters,omitempty"`
	StorageDiff interface{} `json:"storage_diff,omitempty"`
	Mempool     bool        `json:"mempool"`
}

// CodeDiff -
type CodeDiff struct {
	Full    string `json:"full,omitempty"`
	Added   int64  `json:"added,omitempty"`
	Removed int64  `json:"removed,omitempty"`
}

// Contract -
type Contract struct {
	*models.Contract

	Profile *ProfileInfo `json:"profile,omitempty"`
}

// ProfileInfo -
type ProfileInfo struct {
	Subscribed bool `json:"subscribed"`
}

// Subscription -
type Subscription struct {
	*Contract

	SubscribedAt time.Time `json:"subscribed_at"`
}

// TimelineItem -
type TimelineItem struct {
	Event string    `json:"event"`
	Date  time.Time `json:"date"`
}

// OperationResponse -
type OperationResponse struct {
	Operations []Operation `json:"operations"`
	LastID     string      `json:"last_id"`
}

type userProfile struct {
	Login         string         `json:"login"`
	AvatarURL     string         `json:"avatarURL"`
	Subscriptions []Subscription `json:"subscriptions"`
}

// BigMapItem -
type BigMapItem struct {
	Key     interface{} `json:"key"`
	Value   interface{} `json:"value"`
	KeyHash string      `json:"key_hash"`
	Level   int64       `json:"level"`
}

// BigMapResponseItem -
type BigMapResponseItem struct {
	Item  BigMapItem `json:"data"`
	Count int64      `json:"count"`
}
