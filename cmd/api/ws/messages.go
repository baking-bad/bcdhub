package ws

import jsoniter "github.com/json-iterator/go"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Statuses
const (
	ErrorStatus = "error"
	OkStatus    = "ok"
)

// StatusMessage -
type StatusMessage struct {
	Status string `json:"status"`
	Text   string `json:"text"`
}

// ActionMessage -
type ActionMessage struct {
	Action string `json:"action"`
}
