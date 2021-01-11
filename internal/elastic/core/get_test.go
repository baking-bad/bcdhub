package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFiltersToQuery(t *testing.T) {
	tests := []struct {
		name   string
		filter map[string]interface{}
		want   Base
	}{
		{
			name: "test1",
			filter: map[string]interface{}{
				"source.or": []interface{}{
					"tz1W2zByMLGXqemN9jM9s3aagx7cX5S4QojY",
					"tz1Y63jVYqAMTbomAuGadHqohBpDJ95DP1GP",
				},
			},
			want: Base{
				"query": Item{
					"bool": Item{
						"must": []Item{
							{
								"bool": Item{
									"minimum_should_match": 1,
									"should": []Item{
										{
											"match_phrase": Item{
												"source": "tz1W2zByMLGXqemN9jM9s3aagx7cX5S4QojY",
											},
										},
										{
											"match_phrase": Item{
												"source": "tz1Y63jVYqAMTbomAuGadHqohBpDJ95DP1GP",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test2",
			filter: map[string]interface{}{
				".or": []interface{}{
					"tz1W2zByMLGXqemN9jM9s3aagx7cX5S4QojY",
					"tz1Y63jVYqAMTbomAuGadHqohBpDJ95DP1GP",
				},
			},
			want: Base{
				"query": Item{
					"bool": Item{
						"must": []Item{},
					},
				},
			},
		},
		{
			name: "test3",
			filter: map[string]interface{}{
				"source": "tz1W2zByMLGXqemN9jM9s3aagx7cX5S4QojY",
			},
			want: Base{
				"query": Item{
					"bool": Item{
						"must": []Item{
							{
								"match_phrase": Item{
									"source": "tz1W2zByMLGXqemN9jM9s3aagx7cX5S4QojY",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test4",
			filter: map[string]interface{}{
				"source.or": "tz1W2zByMLGXqemN9jM9s3aagx7cX5S4QojY",
			},
			want: Base{
				"query": Item{
					"bool": Item{
						"must": []Item{},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FiltersToQuery(tt.filter)

			if !assert.Equal(t, got, tt.want) {
				t.Errorf("FiltersToQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
