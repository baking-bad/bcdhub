package transfer

import "github.com/baking-bad/bcdhub/internal/elastic/core"

type getTransferedResponse struct {
	Aggs struct {
		Result struct {
			Value struct {
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
	} `json:"aggregations"`
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
