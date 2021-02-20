package types

import "time"

// BigMapDiff -
type BigMapDiff struct {
	Ptr   int64
	Key   []byte
	Value []byte

	ID          string
	KeyHash     string
	OperationID string
	Level       int64
	Address     string
	Network     string
	IndexedTime int64
	Timestamp   time.Time
	Protocol    string
}
