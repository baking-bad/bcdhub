package models

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
)

// SubscriptionRequest -
type SubscriptionRequest struct {
	Address         string
	Network         types.Network
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
	Type      string        `json:"type"`
	Address   string        `json:"address"`
	Network   types.Network `json:"network"`
	Alias     string        `json:"alias"`
	Timestamp time.Time     `json:"-"`
	Body      interface{}   `json:"body,omitempty"`
}

// ByTimestamp - sorting events by timestamp
type ByTimestamp []Event

func (a ByTimestamp) Len() int           { return len(a) }
func (a ByTimestamp) Less(i, j int) bool { return a[i].Timestamp.Before(a[j].Timestamp) }
func (a ByTimestamp) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// ContractCountStats -
type ContractCountStats struct {
	Total     int64
	SameCount int64
}

// NetworkStats -
type NetworkStats struct {
	ContractsCount       uint64
	UniqueContractsCount uint64
	CallsCount           uint64
	FACount              uint64
}

// ContractStats -
type ContractStats struct {
	Count      int64
	LastAction time.Time
}
