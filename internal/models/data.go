package models

import (
	"fmt"
	"time"
)

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
	Type      string      `json:"type"`
	Address   string      `json:"address"`
	Network   string      `json:"network"`
	Alias     string      `json:"alias"`
	Timestamp time.Time   `json:"-"`
	Body      interface{} `json:"body,omitempty"`
}

// ByTimestamp - sorting events by timestamp
type ByTimestamp []Event

func (a ByTimestamp) Len() int           { return len(a) }
func (a ByTimestamp) Less(i, j int) bool { return a[i].Timestamp.Before(a[j].Timestamp) }
func (a ByTimestamp) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// Repository -
type Repository struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// String -
func (repo Repository) String() string {
	return fmt.Sprintf("%s (type: %s)", repo.ID, repo.Type)
}

// ContractCountStats -
type ContractCountStats struct {
	Total     int64
	SameCount int64
}
