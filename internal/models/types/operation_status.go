package types

import "database/sql/driver"

// OperationStatus -
type OperationStatus int

// NewOperationStatus -
func NewOperationStatus(value string) OperationStatus {
	switch value {
	case "applied":
		return OperationStatusApplied
	case "backtracked":
		return OperationStatusBacktracked
	case "failed":
		return OperationStatusFailed
	case "skipped":
		return OperationStatusSkipped
	default:
		return 0
	}
}

// String -
func (status OperationStatus) String() string {
	switch status {
	case OperationStatusApplied:
		return "applied"
	case OperationStatusBacktracked:
		return "backtracked"
	case OperationStatusFailed:
		return "failed"
	case OperationStatusSkipped:
		return "skipped"
	default:
		return ""
	}
}

// Scan -
func (status *OperationStatus) Scan(value interface{}) error {
	*status = OperationStatus(value.(int64))
	return nil
}

// Value -
func (status OperationStatus) Value() (driver.Value, error) { return int(status), nil }

const (
	OperationStatusApplied OperationStatus = iota + 1
	OperationStatusBacktracked
	OperationStatusFailed
	OperationStatusSkipped
)
