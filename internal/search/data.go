package search

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
