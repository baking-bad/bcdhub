package compilation

import (
	"encoding/json"

	"github.com/baking-bad/bcdhub/internal/mq"
)

// Task -
type Task struct {
	ID    uint
	Kind  string
	Files []string
	Dir   string
}

// GetQueues -
func (t Task) GetQueues() []string {
	return []string{mq.QueueCompilations}
}

// MarshalToQueue -
func (t Task) MarshalToQueue() ([]byte, error) {
	return json.Marshal(t)
}
