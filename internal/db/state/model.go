package state

import (
	"time"
)

// State - current indexer state
type State struct {
	ID        int64     `gorm:"AUTO_INCREMENT;unique_index;column:id"`
	Level     int64     `gorm:"level"`
	Timestamp time.Time `gorm:"timestamp"`
	Network   string    `gorm:"network"`
}

// TableName - set table name
func (State) TableName() string {
	return "state"
}
