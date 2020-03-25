package noderpc

import (
	"time"

	"github.com/tidwall/gjson"
)

// Header is a header in a block returned by the Tezos RPC API.
type Header struct {
	Level     int64     `json:"level"`
	Protocol  string    `json:"protocol"`
	Timestamp time.Time `json:"timestamp"`
}

func (h *Header) parseGJSON(data gjson.Result) {
	h.Level = data.Get("level").Int()
	h.Protocol = data.Get("protocol").String()
	h.Timestamp = data.Get("timestamp").Time().UTC()
}
