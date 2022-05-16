package types

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

// lazy storage diff types
const (
	LazyStorageDiffBigMap       = "big_map"
	LazyStorageDiffSaplingState = "sapling_state"
)
