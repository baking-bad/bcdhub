package search

import "fmt"

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
