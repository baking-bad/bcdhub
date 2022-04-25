package types

import "github.com/baking-bad/bcdhub/internal/bcd"

// AccountType -
type AccountType int

// account types
const (
	AccountTypeUnknown = iota
	AccountTypeContract
	AccountTypeTz
)

// NewAccountType -
func NewAccountType(address string) AccountType {
	switch {
	case bcd.IsContract(address):
		return AccountTypeContract
	case bcd.IsAddressLazy(address):
		return AccountTypeTz
	default:
		return AccountTypeUnknown
	}
}
