package models

import "fmt"

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
	Type    string      `json:"type"`
	Address string      `json:"address"`
	Network string      `json:"network"`
	Alias   string      `json:"alias"`
	Body    interface{} `json:"body,omitempty"`
}

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
	Total          int64
	SameCount      int64
	TotalWithdrawn int64
	Balance        int64
}

// Result -
type Result struct {
	Count int64  `json:"count"`
	Time  int64  `json:"time"`
	Items []Item `json:"items"`
}

// Item -
type Item struct {
	Type       string              `json:"type"`
	Value      string              `json:"value"`
	Group      *Group              `json:"group,omitempty"`
	Body       interface{}         `json:"body"`
	Highlights map[string][]string `json:"highlights,omitempty"`

	Network string `json:"-"`
}

// Group -
type Group struct {
	Count int64 `json:"count"`
	Top   []Top `json:"top"`
}

// NewGroup -
func NewGroup(docCount int64) *Group {
	return &Group{
		Count: docCount,
		Top:   make([]Top, 0),
	}
}

// Top -
type Top struct {
	Network string `json:"network"`
	Key     string `json:"key"`
}
