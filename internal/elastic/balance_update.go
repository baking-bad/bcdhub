package elastic

type getBalanceResponse struct {
	Agg struct {
		Balance floatValue `json:"balance"`
	} `json:"aggregations"`
}

// GetBalance -
func (e *Elastic) GetBalance(network, address string) (int64, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				matchPhrase("contract", address),
			),
		),
	).Add(
		aggs(
			aggItem{"balance", sum("change")},
		),
	).Zero()

	var response getBalanceResponse
	if err := e.query([]string{DocBalanceUpdates}, query, &response); err != nil {
		return 0, err
	}
	return int64(response.Agg.Balance.Value), nil
}
