package index

import "time"

// Head -
type Head struct {
	Level     int64
	Hash      string
	Timestamp time.Time
}

// Contract -
type Contract struct {
	Level     int64
	Timestamp time.Time
	Counter   int
	Balance   int64
	Manager   string
	Delegate  string
	Address   string
	Kind      string
}
