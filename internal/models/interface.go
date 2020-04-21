package models

import "github.com/tidwall/gjson"

// Parsable -
type Parsable interface {
	ParseElasticJSON(hit gjson.Result)
}
