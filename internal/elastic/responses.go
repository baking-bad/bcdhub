package elastic

import (
	stdJSON "encoding/json"
	"fmt"
)

// Header -
type Header struct {
	Took     int64 `json:"took"`
	TimedOut bool  `json:"timed_out"`
}

// SearchResponse -
type SearchResponse struct {
	ScrollID string     `json:"_scroll_id,omitempty"`
	Took     int        `json:"took,omitempty"`
	TimedOut *bool      `json:"timed_out,omitempty"`
	Hits     *HitsArray `json:"hits,omitempty"`
}

// HitsArray -
type HitsArray struct {
	Total struct {
		Value    int64  `json:"value"`
		Relation string `json:"relation"`
	} `json:"total"`
	Hits []Hit `json:"hits"`
}

// Hit -
type Hit struct {
	ID        string              `json:"_id"`
	Index     string              `json:"_index"`
	Source    stdJSON.RawMessage  `json:"_source"`
	Score     float64             `json:"_score"`
	Type      string              `json:"_type"`
	Highlight map[string][]string `json:"highlight,omitempty"`
}

// DeleteByQueryResponse -
type DeleteByQueryResponse struct {
	Header
	Total            int64 `json:"total"`
	Deleted          int64 `json:"deleted"`
	VersionConflicts int64 `json:"version_conflicts"`
}

// TestConnectionResponse -
type TestConnectionResponse struct {
	Version struct {
		Number string `json:"number"`
	} `json:"version"`
}

// GetResponse -
type GetResponse struct {
	Index  string             `json:"_index"`
	Type   string             `json:"_type"`
	ID     string             `json:"_id"`
	Found  bool               `json:"found"`
	Source stdJSON.RawMessage `json:"_source"`
}

// BulkResponse -
type BulkResponse struct {
	Took   int64 `json:"took"`
	Errors bool  `json:"errors"`
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

// Bucket -
type Bucket struct {
	Key      string `json:"key"`
	DocCount int64  `json:"doc_count"`
}

type intValue struct {
	Value int64 `json:"value"`
}

type floatValue struct {
	Value float64 `json:"value"`
}

type sqlResponse struct {
	Rows [][]interface{} `json:"rows"`
}
