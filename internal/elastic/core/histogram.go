package core

import "github.com/baking-bad/bcdhub/internal/models/tzip"

// Histogram filter kinds
const (
	HistogramFilterKindExists     = "exists"
	HistogramFilterKindMatch      = "match"
	HistogramFilterKindIn         = "in"
	HistogramFilterKindAddresses  = "address"
	HistogramFilterDexEnrtypoints = "dex_entrypoints"
)

type histogramContext struct {
	Indices  []string
	Period   string
	Function struct {
		Name  string
		Field string
	}
	Filters []HistogramFilter
}

// HistogramFilter -
type HistogramFilter struct {
	Field string
	Value interface{}
	Kind  string
}

func (ctx histogramContext) hasFunction() bool {
	return ctx.Function.Name != "" && ctx.Function.Field != ""
}

func (ctx histogramContext) build() Base {
	hist := Item{
		"date_histogram": Item{
			"field":             "timestamp",
			"calendar_interval": ctx.Period,
		},
	}

	if ctx.hasFunction() {
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
		case HistogramFilterKindExists:
			matches = append(matches, Exists(fltr.Field))
		case HistogramFilterKindMatch:
			matches = append(matches, Match(fltr.Field, fltr.Value))
		case HistogramFilterKindIn:
			if arr, ok := fltr.Value.([]string); ok {
				matches = append(matches, In(fltr.Field, arr))
			}
		case HistogramFilterKindAddresses:
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
		case HistogramFilterDexEnrtypoints:
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
		Aggs(AggItem{"hist", hist}),
	).Zero()
}

// HistogramOption -
type HistogramOption func(*histogramContext)

// WithHistogramIndices -
func WithHistogramIndices(indices ...string) HistogramOption {
	return func(h *histogramContext) {
		h.Indices = indices
	}
}

// WithHistogramFunction -
func WithHistogramFunction(function, field string) HistogramOption {
	return func(h *histogramContext) {
		h.Function = struct {
			Name  string
			Field string
		}{function, field}
	}
}

// WithHistogramFilters -
func WithHistogramFilters(filters []HistogramFilter) HistogramOption {
	return func(h *histogramContext) {
		h.Filters = filters
	}
}

// GetDateHistogram -
func (e *Elastic) GetDateHistogram(period string, opts ...HistogramOption) ([][]int64, error) {
	ctx := histogramContext{
		Period: period,
	}
	for _, opt := range opts {
		opt(&ctx)
	}

	var response getDateHistogramResponse
	if err := e.Query(ctx.Indices, ctx.build(), &response); err != nil {
		return nil, err
	}

	histogram := make([][]int64, 0)
	for _, bucket := range response.Agg.Hist.Buckets {
		val := bucket.DocCount
		if ctx.hasFunction() {
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
