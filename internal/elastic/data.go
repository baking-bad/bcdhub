package elastic

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/models"
)

// LightContract -
type LightContract struct {
	Address  string    `json:"address"`
	Network  string    `json:"network"`
	Deployed time.Time `json:"deploy_time"`
}

// PageableOperations -
type PageableOperations struct {
	Operations []models.Operation `json:"operations"`
	LastID     string             `json:"last_id"`
}

// SameContractsResponse -
type SameContractsResponse struct {
	Count     int64             `json:"count"`
	Contracts []models.Contract `json:"contracts"`
}

// SimilarContract -
type SimilarContract struct {
	*models.Contract
	Count int64 `json:"count"`
}

// BigMapDiff -
type BigMapDiff struct {
	Ptr         int64     `json:"ptr,omitempty"`
	BinPath     string    `json:"bin_path"`
	Key         string    `json:"key"`
	KeyHash     string    `json:"key_hash"`
	Value       string    `json:"value"`
	OperationID string    `json:"operation_id"`
	Level       int64     `json:"level"`
	Address     string    `json:"address"`
	Network     string    `json:"network"`
	Timestamp   time.Time `json:"timestamp"`
	Protocol    string    `json:"protocol"`

	Count int64 `json:"count"`
}

// FromModel -
func (b *BigMapDiff) FromModel(bmd *models.BigMapDiff) error {
	b.Ptr = bmd.Ptr
	b.BinPath = bmd.BinPath
	b.KeyHash = bmd.KeyHash
	b.Value = bmd.Value
	b.OperationID = bmd.OperationID
	b.Level = bmd.Level
	b.Address = bmd.Address
	b.Network = bmd.Network
	b.Timestamp = bmd.Timestamp
	b.Protocol = bmd.Protocol

	bytes, err := json.Marshal(bmd.Key)
	if err != nil {
		return err
	}
	b.Key = string(bytes)
	return nil
}

// ContractStats -
type ContractStats struct {
	TxCount        int64     `json:"tx_count"`
	LastAction     time.Time `json:"last_action"`
	Balance        int64     `json:"balance"`
	TotalWithdrawn int64     `json:"total_withdrawn"`
}

// ContractMigrationsStats -
type ContractMigrationsStats struct {
	MigrationsCount int64 `json:"migrations_count"`
}

// DiffTask -
type DiffTask struct {
	Network1 string
	Address1 string
	Network2 string
	Address2 string
}

// ContractCountStats -
type ContractCountStats struct {
	Total          int64
	SameCount      int64
	TotalWithdrawn int64
	Balance        int64
}

// SubscriptionRequest -
type SubscriptionRequest struct {
	Address         string
	Network         string
	Alias           string
	Hash            string
	ProjectID       string
	WithSame        bool
	WithSimilar     bool
	WithMempool     bool
	WithMigrations  bool
	WithErrors      bool
	WithCalls       bool
	WithDeployments bool
}

// EventContract -
type EventContract struct {
	Network   string    `json:"network"`
	Address   string    `json:"address"`
	Hash      string    `json:"hash"`
	ProjectID string    `json:"project_id"`
	Timestamp time.Time `json:"timestamp"`
}

// EventType -
const (
	EventTypeError     = "error"
	EventTypeMigration = "migration"
	EventTypeCall      = "call"
	EventTypeInvoke    = "invoke"
	EventTypeDeploy    = "deploy"
	EventTypeSame      = "same"
	EventTypeSimilar   = "similar"
	EventTypeMempool   = "mempool"
)

// Event -
type Event struct {
	Type    string      `json:"type"`
	Address string      `json:"address"`
	Network string      `json:"network"`
	Alias   string      `json:"alias"`
	Body    interface{} `json:"body,omitempty"`
}

// EventOperation -
type EventOperation struct {
	Network          string    `json:"network"`
	Hash             string    `json:"hash"`
	Internal         bool      `json:"internal"`
	Status           string    `json:"status"`
	Timestamp        time.Time `json:"timestamp"`
	Kind             string    `json:"kind"`
	Fee              int64     `json:"fee,omitempty"`
	Amount           int64     `json:"amount,omitempty"`
	Entrypoint       string    `json:"entrypoint,omitempty"`
	Source           string    `json:"source"`
	SourceAlias      string    `json:"source_alias,omitempty"`
	Destination      string    `json:"destination,omitempty"`
	DestinationAlias string    `json:"destination_alias,omitempty"`
	Delegate         string    `json:"delegate,omitempty"`
	DelegateAlias    string    `json:"delegate_alias,omitempty"`

	Result *models.OperationResult `json:"result,omitempty"`
	Errors []*cerrors.Error        `json:"errors,omitempty"`
	Burned int64                   `json:"burned,omitempty"`
}

// EventMigration -
type EventMigration struct {
	Network      string    `json:"network"`
	Protocol     string    `json:"protocol"`
	PrevProtocol string    `json:"prev_protocol,omitempty"`
	Hash         string    `json:"hash,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	Level        int64     `json:"level"`
	Address      string    `json:"address"`
	Kind         string    `json:"kind"`
}

// TokenMethodUsageStats -
type TokenMethodUsageStats struct {
	Count       int64
	ConsumedGas int64
}

// TokenUsageStats -
type TokenUsageStats map[string]TokenMethodUsageStats

// DAppStats -
type DAppStats struct {
	Users  int64 `json:"users"`
	Calls  int64 `json:"txs"`
	Volume int64 `json:"volume"`
}

// TransfersResponse -
type TransfersResponse struct {
	Transfers []models.Transfer `json:"transfers"`
	Total     int64             `json:"total"`
	LastID    string            `json:"last_id"`
}

// Address -
type Address struct {
	Address string
	Network string
}
