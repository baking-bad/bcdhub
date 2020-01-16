package models

import "time"

// State -
type State struct {
	ID        string    `json:"-"`
	Level     int64     `json:"level"`
	Timestamp time.Time `json:"timestamp"`
	Network   string    `json:"network"`
	Type      string    `json:"type"`
}
