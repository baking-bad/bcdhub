package types

import (
	stdJSON "encoding/json"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Parameters -
type Parameters struct {
	Entrypoint string             `json:"entrypoint"`
	Value      stdJSON.RawMessage `json:"value"`
}

// NewParameters -
func NewParameters(data []byte) *Parameters {
	var p Parameters
	if err := json.Unmarshal(data, &p); err != nil || p.Entrypoint == "" {
		p.Entrypoint = consts.DefaultEntrypoint
		p.Value = data
	}
	return &p
}
