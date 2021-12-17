package elastic

import (
	stdJSON "encoding/json"
	"time"
)

// TestConnectionResponse -
type TestConnectionResponse struct {
	Version struct {
		Number string `json:"number"`
	} `json:"version"`
}

// Bucket -
type Bucket struct {
	Key      string `json:"key"`
	DocCount uint64 `json:"doc_count"`
}

// IntValue -
type IntValue struct {
	Value int64 `json:"value"`
}

// UintValue -
type UintValue struct {
	Value uint64 `json:"value"`
}

// FloatValue -
type FloatValue struct {
	Value float64 `json:"value"`
}

// TimeValue -
type TimeValue struct {
	Value float64   `json:"value"`
	Time  time.Time `json:"value_as_string"`
}

// SQLResponse -
type SQLResponse struct {
	Rows [][]interface{} `json:"rows"`
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

// HitsArray -
type HitsArray struct {
	Total struct {
		Value    int64  `json:"value"`
		Relation string `json:"relation"`
	} `json:"total"`
	Hits []Hit `json:"hits"`
}

// SearchResponse -
type SearchResponse struct {
	ScrollID string     `json:"_scroll_id,omitempty"`
	Took     int        `json:"took,omitempty"`
	TimedOut *bool      `json:"timed_out,omitempty"`
	Hits     *HitsArray `json:"hits,omitempty"`
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
	Took   int64              `json:"took"`
	Errors bool               `json:"errors"`
	Items  stdJSON.RawMessage `json:"items,omitempty"`
}

// Header -
type Header struct {
	Took     int64 `json:"took"`
	TimedOut bool  `json:"timed_out"`
}

// DeleteByQueryResponse -
type DeleteByQueryResponse struct {
	Header
	Total            int64 `json:"total"`
	Deleted          int64 `json:"deleted"`
	VersionConflicts int64 `json:"version_conflicts"`
}

// searchByTextResponse -
type searchByTextResponse struct {
	Took int64     `json:"took"`
	Hits HitsArray `json:"hits"`
	Agg  struct {
		Projects struct {
			Buckets []struct {
				Bucket
				Last struct {
					Hits HitsArray `json:"hits"`
				} `json:"last"`
			} `json:"buckets"`
		} `json:"projects"`
	} `json:"aggregations"`
}
