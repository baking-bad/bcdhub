package elastic

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/dapp"
)

type getDateHistogramResponse struct {
	Agg struct {
		Hist struct {
			Buckets []struct {
				Key      float64    `json:"key"`
				DocCount float64    `json:"doc_count"`
				Result   FloatValue `json:"result,omitempty"`
			} `json:"buckets"`
		} `json:"hist"`
	} `json:"aggregations"`
}

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
	matches = append(matches, Range("timestamp", Item{"gt": getHistogramInterval(ctx.Period)}))
	for _, fltr := range ctx.Filters {
		switch fltr.Kind {
		case models.HistogramFilterKindExists:
			matches = append(matches, Exists(fltr.Field))
		case models.HistogramFilterKindMatch:
			matches = append(matches, Match(fltr.Field, fltr.Value))
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
			if value, ok := fltr.Value.([]dapp.DAppContract); ok {
				entrypoints := make([]Item, 0)
				for i := range value {
					for j := range value[i].Entrypoint {
						entrypoints = append(entrypoints, Bool(
							Filter(
								MatchPhrase("initiator", value[i].Address),
								Match("parent", value[i].Entrypoint[j]),
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

// Histogram -
func (e *Elastic) Histogram(period string, opts ...models.HistogramOption) ([][]float64, error) {
	ctx := models.HistogramContext{
		Period: period,
	}
	for _, opt := range opts {
		opt(&ctx)
	}

	var response getDateHistogramResponse
	if err := e.query([]string{ctx.Index}, buildHistogramContext(ctx), &response); err != nil {
		return nil, err
	}

	histogram := make([][]float64, 0)
	for _, bucket := range response.Agg.Hist.Buckets {
		val := bucket.DocCount
		if ctx.HasFunction() {
			val = bucket.Result.Value
		}

		item := []float64{
			bucket.Key,
			val,
		}
		histogram = append(histogram, item)
	}
	return histogram, nil
}

func getHistogramInterval(period string) string {
	switch period {
	case "hour":
		return "now-1d"
	case "day":
		return "now-1M"
	case "week":
		return "now-16w"
	case "month":
		return "now-1y"
	default:
		return "2018-06-25"
	}
}
