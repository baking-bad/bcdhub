package transfer

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
	} `json:"aggregations"`
}
