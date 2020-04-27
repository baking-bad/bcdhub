package models

import "github.com/tidwall/gjson"

// Parsable -
type Parsable interface {
	ParseElasticJSON(hit gjson.Result)
}

// Scorable -
type Scorable interface {
	GetScores(search string) []string
	FoundByName(hit gjson.Result) string
}
