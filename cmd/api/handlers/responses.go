package handlers

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/baking-bad/bcdhub/internal/models"
)

// Operation -
type Operation struct {
	ID        string    `json:"id"`
	Protocol  string    `json:"protocol"`
	Hash      string    `json:"hash"`
	Internal  bool      `json:"internal"`
	Network   string    `json:"network"`
	Timesatmp time.Time `json:"timestamp"`

	Level            int64            `json:"level"`
	Kind             string           `json:"kind"`
	Source           string           `json:"source"`
	SourceAlias      string           `json:"source_alias,omitempty"`
	Fee              int64            `json:"fee,omitempty"`
	Counter          int64            `json:"counter,omitempty"`
	GasLimit         int64            `json:"gas_limit,omitempty"`
	StorageLimit     int64            `json:"storage_limit,omitempty"`
	Amount           int64            `json:"amount,omitempty"`
	Destination      string           `json:"destination,omitempty"`
	DestinationAlias string           `json:"destination_alias,omitempty"`
	PublicKey        string           `json:"public_key,omitempty"`
	ManagerPubKey    string           `json:"manager_pubkey,omitempty"`
	Balance          int64            `json:"balance,omitempty"`
	Delegate         string           `json:"delegate,omitempty"`
	Status           string           `json:"status"`
	Entrypoint       string           `json:"entrypoint,omitempty"`
	Errors           []cerrors.IError `json:"errors,omitempty"`
	Burned           int64            `json:"burned,omitempty"`

	BalanceUpdates []models.BalanceUpdate  `json:"balance_updates,omitempty"`
	Result         *models.OperationResult `json:"result,omitempty"`

	Parameters  interface{} `json:"parameters,omitempty"`
	StorageDiff interface{} `json:"storage_diff,omitempty"`
	Mempool     bool        `json:"mempool"`

	IndexedTime int64 `json:"-"`
}

// Contract -
type Contract struct {
	*models.Contract

	Profile *ProfileInfo `json:"profile,omitempty"`
	Slug    string       `json:"slug,omitempty"`
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
	Key       interface{} `json:"key"`
	Value     interface{} `json:"value"`
	KeyHash   string      `json:"key_hash"`
	KeyString string      `json:"key_string"`
	Level     int64       `json:"level"`
	Timestamp time.Time   `json:"timestamp"`
}

// BigMapResponseItem -
type BigMapResponseItem struct {
	Item  BigMapItem `json:"data"`
	Count int64      `json:"count"`
}

// Migration -
type Migration struct {
	Level        int64     `json:"level"`
	Timestamp    time.Time `json:"timestamp"`
	Hash         string    `json:"hash,omitempty"`
	Protocol     string    `json:"protocol"`
	PrevProtocol string    `json:"prev_protocol"`
	Kind         string    `json:"kind"`
}

// TokenContract -
type TokenContract struct {
	Network       string    `json:"network"`
	Level         int64     `json:"level"`
	Timestamp     time.Time `json:"timestamp"`
	Address       string    `json:"address"`
	Manager       string    `json:"manager,omitempty"`
	Delegate      string    `json:"delegate,omitempty"`
	Alias         string    `json:"alias,omitempty"`
	DelegateAlias string    `json:"delegate_alias,omitempty"`
	Type          string    `json:"type"`
	Balance       int64     `json:"balance"`
	TxCount       int64     `json:"tx_count,omitempty"`
}

// TokenTransfer -
type TokenTransfer struct {
	Contract  string    `json:"contract"`
	Network   string    `json:"network"`
	Protocol  string    `json:"protocol"`
	Hash      string    `json:"hash"`
	Counter   int64     `json:"counter,omitempty"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Level     int64     `json:"level"`
	From      string    `json:"from,omitempty"`
	To        string    `json:"to"`
	Amount    int64     `json:"amount"`
	Source    string    `json:"source"`
}

// PageableTokenTransfers -
type PageableTokenTransfers struct {
	Transfers []TokenTransfer `json:"transfers"`
	LastID    string          `json:"last_id"`
}

// BigMapDiffItem -
type BigMapDiffItem struct {
	Value     interface{} `json:"value"`
	Level     int64       `json:"level"`
	Timestamp time.Time   `json:"timestamp"`
}

// BigMapDiffByKeyResponse -
type BigMapDiffByKeyResponse struct {
	Key     interface{}      `json:"key,omitempty"`
	KeyHash string           `json:"key_hash"`
	Values  []BigMapDiffItem `json:"values,omitempty"`
	Total   int64            `json:"total"`
}

// CodeDiffResponse -
type CodeDiffResponse struct {
	Left  CodeDiffLeg          `json:"left"`
	Right CodeDiffLeg          `json:"right"`
	Diff  formatter.DiffResult `json:"diff"`
}

// NetworkStats -
type NetworkStats struct {
	ContractsCount  int64             `json:"contracts_count"`
	OperationsCount int64             `json:"operations_count"`
	Protocols       []models.Protocol `json:"protocols"`
}

// SearchBigMapDiff -
type SearchBigMapDiff struct {
	Ptr       int64     `json:"ptr"`
	Key       string    `json:"key"`
	KeyHash   string    `json:"key_hash"`
	Value     string    `json:"value"`
	Level     int64     `json:"level"`
	Address   string    `json:"address"`
	Network   string    `json:"network"`
	Timestamp time.Time `json:"timestamp"`
	FoundBy   string    `json:"found_by"`
}

// GetErrorLocationResponse -
type GetErrorLocationResponse struct {
	Text        string `json:"text"`
	FailedRow   int    `json:"failed_row"`
	FirstRow    int    `json:"first_row"`
	StartColumn int    `json:"start_col"`
	EndColumn   int    `json:"end_col"`
}
