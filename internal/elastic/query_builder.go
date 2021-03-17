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

// MinimumShouldMatch -
func MinimumShouldMatch(value int) Item {
	return Item{
		"minimum_should_match": value,
	}
}

// Exists -
func Exists(field string) Item {
	return Item{
		"exists": Item{
			"field": field,
		},
	}
}

// Must -
func Must(items ...Item) Item {
	return Item{
		"must": items,
	}
}

// MustNot -
func MustNot(items ...Item) Item {
	return Item{
		"must_not": items,
	}
}

// Should -
func Should(items ...Item) Item {
	return Item{
		"should": items,
	}
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

// Term -
func Term(key string, value interface{}) Item {
	return Item{
		"term": Item{
			key: value,
		},
	}
}

// In -
func In(key string, value []string) Item {
	return Item{
		"terms": Item{
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

// Cardinality -
func Cardinality(field string) Item {
	return Item{
		"cardinality": Item{
			"field": field,
		},
	}
}

// Avg -
func Avg(field string) Item {
	return Item{
		"avg": Item{
			"field": field,
		},
	}
}

// TermsAgg -
func TermsAgg(field string, size int64) Item {
	t := Item{
		"field": field,
	}
	if size > 0 {
		t["size"] = size
	}
	return Item{
		"terms": t,
	}
}

// Composite -
func Composite(size int64, items ...AggItem) Item {
	body := make([]Item, 0)
	for i := range items {
		body = append(body, Item{
			items[i].Name: items[i].Body,
		})
	}
	return Item{
		"composite": Item{
			"sources": body,
			"size":    size,
		},
	}
}

// TopHits -
func TopHits(size int, sortField, order string) Item {
	return Item{
		"top_hits": Item{
			"size": size,
			"sort": Sort(sortField, order),
		},
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

// Max -
func Max(field string) Item {
	return Item{
		"max": Item{
			"field": field,
		},
	}
}

// Min -
func Min(field string) Item {
	return Item{
		"min": Item{
			"field": field,
		},
	}
}

// Sum -
func Sum(field string) Item {
	return Item{
		"sum": Item{
			"field": field,
		},
	}
}

// Count -
func Count(field string) Item {
	return Item{
		"value_count": Item{
			"field": field,
		},
	}
}

// MaxBucket -
func MaxBucket(bucketsPath string) Item {
	return Item{
		"max_bucket": Item{
			"buckets_path": bucketsPath,
		},
	}
}

// MinBucket -
func MinBucket(bucketsPath string) Item {
	return Item{
		"min_bucket": Item{
			"buckets_path": bucketsPath,
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

// Append -
func (q Item) Append(key string, value interface{}) Item {
	q[key] = value
	return q
}

// Extend -
func (q Item) Extend(item Item) Item {
	for k, v := range item {
		q[k] = v
	}
	return q
}

// Get -
func (q Item) Get(name string) Item {
	if val, ok := q[name]; ok {
		if typ, ok := val.(Item); ok {
			return typ
		}
		return nil
	}
	return nil
}

// Base -
type Base Item

// NewQuery -
func NewQuery() Base {
	return Base{}
}

// Size -
func (q Base) Size(size int64) Base {
	if size != 0 {
		q["size"] = size
	}
	return q
}

// All -
func (q Base) All() Base {
	q["size"] = MaxQuerySize
	return q
}

// One -
func (q Base) One() Base {
	q["size"] = 1
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

// SearchAfter -
func (q Base) SearchAfter(value []interface{}) Base {
	q["search_after"] = value
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

// Source -
func (q Base) Source(items ...Item) Base {
	qi := Item{}
	for i := range items {
		for k, v := range items[i] {
			if helpers.StringInArray(k, []string{"excludes", "includes"}) {
				qi[k] = v
			}
		}
	}
	q["_source"] = qi
	return q
}

// Highlights -
func (q Base) Highlights(highlights Item) Base {
	q["highlight"] = Item{
		"fields": highlights,
	}
	return q
}

// Get -
func (q Base) Get(name string) Item {
	if val, ok := q[name]; ok {
		if typ, ok := val.(Item); ok {
			return typ
		}
		return nil
	}
	return nil
}
