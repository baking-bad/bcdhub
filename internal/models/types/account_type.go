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
	AccountTypeSmartRollup
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
	case bcd.IsSmartRollupAddressLazy(address):
		return AccountTypeSmartRollup
	default:
		return AccountTypeUnknown
	}
}

// String -
func (typ AccountType) String() string {
	switch typ {
	case AccountTypeContract:
		return "contract"
	case AccountTypeRollup:
		return "rollup"
	case AccountTypeSmartRollup:
		return "smart_rollup"
	case AccountTypeTz:
		return "account"
	default:
		return "unknown"
	}
}
