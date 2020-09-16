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

// GetQueue -
func (t Task) GetQueue() string {
	return mq.QueueCompilations
}

// Marshal -
func (t Task) Marshal() ([]byte, error) {
	return json.Marshal(t)
}
