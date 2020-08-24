package elastic

type histogramContext struct {
	Indices  []string
	Period   string
	Function struct {
		Name  string
		Field string
	}
	Addresses []string
	Filters   map[string]interface{}
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
		hist.Append("aggs", qItem{
			"result": qItem{
				ctx.Function.Name: qItem{
					"field": ctx.Function.Field,
				},
			},
		})
	}

	matches := make([]qItem, 0)
	for key, value := range ctx.Filters {
		if value == "" {
			matches = append(matches, exists(key))
		} else {
			matches = append(matches, matchQ(key, value))
		}
	}

	if len(ctx.Addresses) > 0 {
		addresses := make([]qItem, len(ctx.Addresses))
		for i := range ctx.Addresses {
			addresses[i] = matchPhrase("destination", ctx.Addresses[i])
		}
		matches = append(matches, boolQ(
			should(addresses...),
			minimumShouldMatch(1),
		))
	}

	return newQuery().Query(
		boolQ(
			filter(
				matches...,
			),
		),
	).Add(
		aggs("hist", hist),
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

// WithHistogramAddresses -
func WithHistogramAddresses(addresses ...string) HistogramOption {
	return func(h *histogramContext) {
		h.Addresses = addresses
	}
}

// WithHistogramFilters -
func WithHistogramFilters(filters map[string]interface{}) HistogramOption {
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
