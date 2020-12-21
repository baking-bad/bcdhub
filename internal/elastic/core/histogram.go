package core

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

func buildHistogramContext(ctx models.HistogramContext) Base {
	hist := Item{
		"date_histogram": Item{
			"field":             "timestamp",
			"calendar_interval": ctx.Period,
		},
	}

	if ctx.HasFunction() {
		hist.Extend(Aggs(
			AggItem{
				"result", Item{
					ctx.Function.Name: Item{
						"field": ctx.Function.Field,
					},
				},
			},
		))
	}

	matches := make([]Item, 0)
	for _, fltr := range ctx.Filters {
		switch fltr.Kind {
		case models.HistogramFilterKindExists:
			matches = append(matches, Exists(fltr.Field))
		case models.HistogramFilterKindMatch:
			matches = append(matches, Match(fltr.Field, fltr.Value))
		case models.HistogramFilterKindIn:
			if arr, ok := fltr.Value.([]string); ok {
				matches = append(matches, In(fltr.Field, arr))
			}
		case models.HistogramFilterKindAddresses:
			if value, ok := fltr.Value.([]string); ok {
				addresses := make([]Item, len(value))
				for i := range value {
					addresses[i] = MatchPhrase(fltr.Field, value[i])
				}
				matches = append(matches, Bool(
					Should(addresses...),
					MinimumShouldMatch(1),
				))
			}
		case models.HistogramFilterDexEnrtypoints:
			if value, ok := fltr.Value.([]tzip.DAppContract); ok {
				entrypoints := make([]Item, 0)
				for i := range value {
					for j := range value[i].DexVolumeEntrypoints {
						entrypoints = append(entrypoints, Bool(
							Filter(
								MatchPhrase("initiator", value[i].Address),
								Match("parent", value[i].DexVolumeEntrypoints[j]),
							),
						))
					}
				}
				matches = append(matches, Bool(
					Should(entrypoints...),
					MinimumShouldMatch(1),
				))
			}
		}
	}

	return NewQuery().Query(
		Bool(
			Filter(
				matches...,
			),
		),
	).Add(
		Aggs(AggItem{Name: "hist", Body: hist}),
	).Zero()
}

// GetDateHistogram -
func (e *Elastic) GetDateHistogram(period string, opts ...models.HistogramOption) ([][]int64, error) {
	ctx := models.HistogramContext{
		Period: period,
	}
	for _, opt := range opts {
		opt(&ctx)
	}

	var response getDateHistogramResponse
	if err := e.Query(ctx.Indices, buildHistogramContext(ctx), &response); err != nil {
		return nil, err
	}

	histogram := make([][]int64, 0)
	for _, bucket := range response.Agg.Hist.Buckets {
		val := bucket.DocCount
		if ctx.HasFunction() {
			val = int64(bucket.Result.Value)
		}

		item := []int64{
			bucket.Key,
			val,
		}
		histogram = append(histogram, item)
	}
	return histogram, nil
}
