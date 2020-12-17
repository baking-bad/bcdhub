package events

import "errors"

// Names
const (
	SingleAssetBalanceUpdates = "singleassetbalanceupdates"
	MultiAssetBalanceUpdates  = "multiassetbalanceupdates"
)

// errors
var (
	ErrNodeReturn = errors.New(`Node return error`)
)
