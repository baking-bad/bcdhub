package elastic

import	"github.com/aopoltorzhicky/bcdhub/internal/helpers"

type qItem map[string]interface{}

func boolQ(items... qItem) qItem {
	bq := qItem{}
	q := qItem{}
	for i := range items {
		for k, v := range items[i] {
			if helpers.StringInArray(k, []string{"must", "should", "filter", "not_must", "minimum_should_match"}) {
				q[k] = v
			}
		}
	}
	bq["bool"] = q
	return bq
}

func must(items... qItem) qItem {
	return qItem{
		"must": items,
	}
}

func notMust(items... qItem) qItem {
	return qItem{
		"not_must": items,
	}
}

func should(items... qItem) qItem {
	return qItem{
		"should": items,
	}
}

func filter(items... qItem) qItem {
	return qItem{
		"filter": items,
	}
}

func matchPhrase(key string, value interface{}) qItem {
	return qItem{
		"match_phrase": qItem {
			key: value,
		},
	}
}

func term(key string, value interface{}) qItem {
	return qItem{
		"term": qItem {
			key: value,
		},
	}
}

func excludes(fields []string) qItem {
	return qItem{
		"excludes": fields,
	}
}

func includes(fields []string) qItem {
	return qItem{
		"includes": fields,
	}
}

type base qItem

func newQuery() base {
	return base{}
}

func (q base) Size(size int) base {
	q["size"] = size
	return q
}

func (q base) All() base {
	q["size"] = 10000
	return q
}

func (q base) One() base {
	q["size"] = 1
	return q
}

func (q base) Zero() base {
	q["size"] = 0
	return q
}

func (q base) From(from int) base {
	q["from"] = from
	return q
}

func (q base) Query(item qItem) base {
	q["query"] = item
	return q
}

func (q base) Sort(key, order string) base {
	q["sort"] = qItem{
		key: qItem {
			"order": order,
		},
	}
	return q
}

func (q base) Source(items... qItem) base {
	qi := qItem{}
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
