package transfer

import "github.com/baking-bad/bcdhub/internal/elastic/core"

type getAccountBalancesResponse struct {
	Agg struct {
		Balances struct {
			Value map[string]float64 `json:"value"`
		} `json:"balances"`
	} `json:"aggregations"`
}

type getTokenSupplyResponse struct {
	Aggs struct {
		Result struct {
			Value struct {
				Supply     float64 `json:"supply"`
				Transfered float64 `json:"transfered"`
			} `json:"value"`
		} `json:"result"`
	} `json:"aggregations"`
}

type aggVolumeSumResponse struct {
	Aggs struct {
		Result struct {
			Value float64 `json:"value"`
		} `json:"volume"`
	}
}

type getTokenVolumeSeriesResponse struct {
	Agg struct {
		Hist struct {
			Buckets []struct {
				Key    int64           `json:"key"`
				Result core.FloatValue `json:"result"`
			} `json:"buckets"`
		} `json:"hist"`
	} `json:"aggregations"`
}
