package bigmapdiff

import "github.com/baking-bad/bcdhub/internal/elastic/core"

type getBigMapDiffsWithKeysResponse struct {
	Agg struct {
		Keys struct {
			Buckets []struct {
				DocCount int64 `json:"doc_count"`
				TopKey   struct {
					Hits core.HitsArray `json:"hits"`
				} `json:"top_key"`
			} `json:"buckets"`
		} `json:"keys"`
	} `json:"aggregations"`
}

type getBigMapDiffsCountResponse struct {
	Agg struct {
		Count core.IntValue `json:"count"`
	} `json:"aggregations"`
}
