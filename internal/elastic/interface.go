package elastic

import "github.com/tidwall/gjson"

// Model -
type Model interface {
	GetID() string
	GetIndex() string
	ParseElasticJSON(gjson.Result)
}

// Scorable -
type Scorable interface {
	GetScores(string) []string
	FoundByName(gjson.Result) string
}
