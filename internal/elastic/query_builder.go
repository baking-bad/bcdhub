package elastic

import (
	"github.com/baking-bad/bcdhub/internal/helpers"
)

// sizes
const (
	MaxQuerySize = 10000
	MinQuerySize = 0
)

// Item -
type Item map[string]interface{}

// List -
type List []interface{}

// Bool -
func Bool(items ...Item) Item {
	bq := Item{}
	q := Item{}
	for i := range items {
		for k, v := range items[i] {
			if helpers.StringInArray(k, []string{"must", "should", "filter", "must_not", "minimum_should_match"}) {
				q[k] = v
			}
		}
	}
	bq["bool"] = q
	return bq
}

// Filter -
func Filter(items ...Item) Item {
	return Item{
		"filter": items,
	}
}

// Range -
func Range(field string, orders ...Item) Item {
	q := Item{}
	for i := range orders {
		for k, v := range orders[i] {
			if helpers.StringInArray(k, []string{"lt", "gt", "lte", "gte"}) {
				q[k] = v
			}
		}
	}
	return Item{
		"range": Item{
			field: q,
		},
	}
}

// MatchPhrase -
func MatchPhrase(key string, value interface{}) Item {
	return Item{
		"match_phrase": Item{
			key: value,
		},
	}
}

// Match -
func Match(key string, value interface{}) Item {
	return Item{
		"match": Item{
			key: value,
		},
	}
}

// AggItem -
type AggItem struct {
	Name string
	Body Item
}

// Aggs -
func Aggs(items ...AggItem) Item {
	body := Item{}
	for i := range items {
		body[items[i].Name] = items[i].Body
	}
	return Item{
		"aggs": body,
	}
}

// Sort -
func Sort(field, order string) Item {
	return Item{
		field: Item{
			"order": order,
		},
	}
}

// QueryString -
func QueryString(text string, fields []string) Item {
	queryS := Item{
		"query": text,
	}
	if len(fields) > 0 {
		queryS["fields"] = fields
	}
	return Item{
		"query_string": queryS,
	}
}

// Base -
type Base Item

// NewQuery -
func NewQuery() Base {
	return Base{}
}

// All -
func (q Base) All() Base {
	q["size"] = MaxQuerySize
	return q
}

// Zero -
func (q Base) Zero() Base {
	q["size"] = MinQuerySize
	return q
}

// From -
func (q Base) From(from int64) Base {
	if from != 0 {
		q["from"] = from
	}
	return q
}

// Query -
func (q Base) Query(item Item) Base {
	q["query"] = item
	return q
}

// Sort -
func (q Base) Sort(key, order string) Base {
	q["sort"] = Item{
		key: Item{
			"order": order,
		},
	}
	return q
}

// Add -
func (q Base) Add(items ...Item) Base {
	for _, item := range items {
		for k, v := range item {
			q[k] = v
		}
	}
	return q
}

// Highlights -
func (q Base) Highlights(highlights Item) Base {
	q["highlight"] = Item{
		"fields": highlights,
	}
	return q
}
