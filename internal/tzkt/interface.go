package tzkt

import "github.com/tidwall/gjson"

// Service -
type Service interface {
	GetMempool(address string) (gjson.Result, error)
}
