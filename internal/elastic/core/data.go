package core

import (
	stdJSON "encoding/json"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/models/operation"
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
	DocCount int64  `json:"doc_count"`
}

// IntValue -
type IntValue struct {
	Value int64 `json:"value"`
}

// FloatValue -
type FloatValue struct {
	Value float64 `json:"value"`
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
	Took   int64 `json:"took"`
	Errors bool  `json:"errors"`
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

	Result *operation.Result  `json:"result,omitempty"`
	Errors []*tezerrors.Error `json:"errors,omitempty"`
	Burned int64              `json:"burned,omitempty"`
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

// EventContract -
type EventContract struct {
	Network   string    `json:"network"`
	Address   string    `json:"address"`
	Hash      string    `json:"hash"`
	ProjectID string    `json:"project_id"`
	Timestamp time.Time `json:"timestamp"`
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

type getDateHistogramResponse struct {
	Agg struct {
		Hist struct {
			Buckets []struct {
				Key      int64      `json:"key"`
				DocCount int64      `json:"doc_count"`
				Result   FloatValue `json:"result,omitempty"`
			} `json:"buckets"`
		} `json:"hist"`
	} `json:"aggregations"`
}
