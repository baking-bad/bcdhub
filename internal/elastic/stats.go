package elastic

import (
	"fmt"
)

// GetItemsCountForNetwork -
func (e *Elastic) GetItemsCountForNetwork(network string) (stats NetworkCountStats, err error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
			),
		),
	).Add(
		aggs("by_index", qItem{
			"terms": qItem{
				"field": "_index",
				"size":  maxQuerySize,
			},
		}),
	).Zero()

	response, err := e.query([]string{DocContracts, DocOperations}, query)
	if err != nil {
		return
	}

	data := response.Get("aggregations.by_index.buckets").Array()
	for _, item := range data {
		key := item.Get("key").String()
		count := item.Get("doc_count").Int()
		switch key {
		case DocContracts:
			stats.Contracts = count
		case DocOperations:
			stats.Operations = count
		default:
			return stats, fmt.Errorf("Unknwon index: %s", key)
		}
	}

	return
}

// GetDateHistogram -
func (e *Elastic) GetDateHistogram(network, index, period string) ([][]int64, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
			),
		),
	).Add(
		aggs("hist", qItem{
			"date_histogram": qItem{
				"field":             "timestamp",
				"calendar_interval": period,
			},
		}),
	).Zero()

	response, err := e.query([]string{index}, query)
	if err != nil {
		return nil, err
	}

	data := response.Get("aggregations.hist.buckets").Array()
	histogram := make([][]int64, 0)
	for _, hit := range data {
		item := []int64{
			hit.Get("key").Int(),
			hit.Get("doc_count").Int(),
		}
		histogram = append(histogram, item)
	}
	return histogram, nil
}
