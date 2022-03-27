package search

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models/types"
)

// Result -
type Result struct {
	Count int64   `json:"count"`
	Time  int64   `json:"time"`
	Items []*Item `json:"items"`
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
	Count uint64 `json:"count"`
	Top   []Top  `json:"top"`
}

// NewGroup -
func NewGroup(docCount uint64) *Group {
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

// Repository -
type Repository struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// String -
func (repo Repository) String() string {
	return fmt.Sprintf("%s (type: %s)", repo.ID, repo.Type)
}

// BigMapDiffSearchArgs -
type BigMapDiffSearchArgs struct {
	Network  types.Network
	Contract string
	Ptr      *int64
	Query    string
	Size     int64
	Offset   int64
	MinLevel *int64
	MaxLevel *int64
}

// BigMapDiffResult -
type BigMapDiffResult struct {
	Key   string `json:"key"`
	Count int64  `json:"doc_count"`
}
