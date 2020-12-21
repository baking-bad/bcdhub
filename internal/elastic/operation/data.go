package operation

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/elastic/core"
)

type getOperationsStatsResponse struct {
	Aggs struct {
		OPG struct {
			Value int64 `json:"value"`
		} `json:"opg"`
		LastAction struct {
			Value time.Time `json:"value_as_string"`
		} `json:"last_action"`
	} `json:"aggregations"`
}

type getByContract struct {
	Hist core.HitsArray `json:"hits"`
	Agg  struct {
		LastID core.FloatValue `json:"last_id"`
	} `json:"aggregations"`
}

type aggVolumeSumResponse struct {
	Aggs struct {
		Result struct {
			Value float64 `json:"value"`
		} `json:"volume"`
	}
}

type getTokensStatsResponse struct {
	Aggs struct {
		Body struct {
			Buckets []struct {
				DocCount int64 `json:"doc_count"`
				Key      struct {
					Destination string `json:"destination"`
					Entrypoint  string `json:"entrypoint"`
				} `json:"key"`
				AVG core.FloatValue `json:"average_consumed_gas"`
			} `json:"buckets"`
		} `json:"body"`
	} `json:"aggregations"`
}
