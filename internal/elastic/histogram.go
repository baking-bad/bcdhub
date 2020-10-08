package elastic

// Histogram filter kinds
const (
	HistogramFilterKindExists    = "exists"
	HistogramFilterKindMatch     = "match"
	HistogramFilterKindIn        = "in"
	HistogramFilterKindAddresses = "address"
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

func (ctx histogramContext) build() base {
	hist := qItem{
		"date_histogram": qItem{
			"field":             "timestamp",
			"calendar_interval": ctx.Period,
		},
	}

	if ctx.hasFunction() {
		hist.Extend(aggs(
			aggItem{
				"result", qItem{
					ctx.Function.Name: qItem{
						"field": ctx.Function.Field,
					},
				},
			},
		))
	}

	matches := make([]qItem, 0)
	for _, filter := range ctx.Filters {
		switch filter.Kind {
		case HistogramFilterKindExists:
			matches = append(matches, exists(filter.Field))
		case HistogramFilterKindMatch:
			matches = append(matches, matchQ(filter.Field, filter.Value))
		case HistogramFilterKindIn:
			if arr, ok := filter.Value.([]string); ok {
				matches = append(matches, in(filter.Field, arr))
			}
		case HistogramFilterKindAddresses:
			if value, ok := filter.Value.([]string); ok {
				addresses := make([]qItem, len(value))
				for i := range value {
					addresses[i] = matchPhrase(filter.Field, value[i])
				}
				matches = append(matches, boolQ(
					should(addresses...),
					minimumShouldMatch(1),
				))
			}
		}
	}

	return newQuery().Query(
		boolQ(
			filter(
				matches...,
			),
		),
	).Add(
		aggs(aggItem{"hist", hist}),
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

	response, err := e.query(ctx.Indices, ctx.build())
	if err != nil {
		return nil, err
	}

	data := response.Get("aggregations.hist.buckets").Array()
	histogram := make([][]int64, 0)
	for _, hit := range data {
		key := "doc_count"
		if ctx.hasFunction() {
			key = "result.value"
		}

		item := []int64{
			hit.Get("key").Int(),
			hit.Get(key).Int(),
		}
		histogram = append(histogram, item)
	}
	return histogram, nil
}
