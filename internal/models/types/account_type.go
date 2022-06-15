package types

import "github.com/baking-bad/bcdhub/internal/bcd"

// AccountType -
type AccountType int

// account types
const (
	AccountTypeUnknown = iota
	AccountTypeContract
	AccountTypeTz
	AccountTypeRollup
)

// NewAccountType -
func NewAccountType(address string) AccountType {
	switch {
	case bcd.IsContract(address):
		return AccountTypeContract
	case bcd.IsAddressLazy(address):
		return AccountTypeTz
	case bcd.IsRollupAddressLazy(address):
		return AccountTypeRollup
	default:
		return AccountTypeUnknown
	}
}
