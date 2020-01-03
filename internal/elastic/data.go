package elastic

import "time"

// InfoResponse -
type InfoResponse struct {
	Name        string `json:"name"`
	ClusterName string `json:"cluster_name"`
	ClusterUUID string `json:"cluster_uuid"`
	Version     struct {
		Number                           string    `json:"number"`
		BuildFlavor                      string    `json:"build_flavor"`
		BuildType                        string    `json:"build_type"`
		BuildHash                        string    `json:"build_hash"`
		BuildDate                        time.Time `json:"build_date"`
		BuildSnapshot                    bool      `json:"build_snapshot"`
		LuceneVersion                    string    `json:"lucene_version"`
		MinimumWireCompatibilityVersion  string    `json:"minimum_wire_compatibility_version"`
		MinimumIndexCompatibilityVersion string    `json:"minimum_index_compatibility_version"`
	} `json:"version"`
	Tagline string `json:"tagline"`
}

// DefaultResponse -
type DefaultResponse struct {
	Index       string                 `json:"_index"`
	Type        string                 `json:"_type"`
	ID          string                 `json:"_id"`
	Version     int                    `json:"_version,omitempty"`
	SeqNo       int                    `json:"_seq_no,omitempty"`
	PrimaryTerm int                    `json:"_primary_term,omitempty"`
	Found       bool                   `json:"found,omitempty"`
	Score       float64                `json:"_score,omitempty"`
	Source      map[string]interface{} `json:"_source"`
}

// SearchResponse -
type SearchResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits Hit `json:"hits"`
}

// Hit -
type Hit struct {
	Total struct {
		Value    int    `json:"value"`
		Relation string `json:"relation"`
	} `json:"total"`
	MaxScore float64           `json:"max_score"`
	Hits     []DefaultResponse `json:"hits"`
}
