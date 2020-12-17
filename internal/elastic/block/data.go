package block

import "github.com/baking-bad/bcdhub/internal/elastic/core"

type getLastBlocksResponse struct {
	Agg struct {
		ByNetwork struct {
			Buckets []struct {
				Last struct {
					Hits core.HitsArray `json:"hits"`
				} `json:"last"`
			} `json:"buckets"`
		} `json:"by_network"`
	} `json:"aggregations"`
}
