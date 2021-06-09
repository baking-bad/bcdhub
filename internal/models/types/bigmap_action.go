package types

import "database/sql/driver"

// BigMapAction -
type BigMapAction int

// NewBigMapAction -
func NewBigMapAction(value string) BigMapAction {
	switch value {
	case BigMapActionStringAlloc:
		return BigMapActionAlloc
	case BigMapActionStringUpdate:
		return BigMapActionUpdate
	case BigMapActionStringCopy:
		return BigMapActionCopy
	case BigMapActionStringRemove:
		return BigMapActionRemove
	default:
		return 0
	}
}

// String -
func (action BigMapAction) String() string {
	switch action {
	case BigMapActionAlloc:
		return BigMapActionStringAlloc
	case BigMapActionUpdate:
		return BigMapActionStringUpdate
	case BigMapActionCopy:
		return BigMapActionStringCopy
	case BigMapActionRemove:
		return BigMapActionStringRemove
	default:
		return ""
	}
}

// Scan -
func (action *BigMapAction) Scan(value interface{}) error {
	*action = BigMapAction(value.(int64))
	return nil
}

// Value -
func (action BigMapAction) Value() (driver.Value, error) { return int(action), nil }

// int values
const (
	BigMapActionAlloc BigMapAction = iota + 1
	BigMapActionUpdate
	BigMapActionCopy
	BigMapActionRemove
)

// string values
const (
	BigMapActionStringUpdate = "update"
	BigMapActionStringCopy   = "copy"
	BigMapActionStringAlloc  = "alloc"
	BigMapActionStringRemove = "remove"
)
