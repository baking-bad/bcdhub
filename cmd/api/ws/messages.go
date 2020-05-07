package ws

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
