package contract

import "github.com/baking-bad/bcdhub/internal/elastic/core"

type getDiffTasksResponse struct {
	Agg struct {
		Projects struct {
			Buckets []struct {
				core.Bucket
				Last struct {
					Hits core.HitsArray `json:"hits"`
				} `json:"last"`
				ByHash struct {
					Buckets []struct {
						core.Bucket
						Last struct {
							Hits core.HitsArray `json:"hits"`
						} `json:"last"`
					} `json:"buckets"`
				} `json:"by_hash"`
			} `json:"buckets"`
		} `json:"by_project"`
	} `json:"aggregations"`
}

type getProjectsResponse struct {
	Agg struct {
		Projects struct {
			Buckets []struct {
				core.Bucket
				Last struct {
					Hits core.HitsArray `json:"hits"`
				} `json:"last"`
			} `json:"buckets"`
		} `json:"projects"`
	} `json:"aggregations"`
}
