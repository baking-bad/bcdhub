package types

import "database/sql/driver"

// OperationKind -
type OperationKind int

// NewOperationKind -
func NewOperationKind(value string) OperationKind {
	switch value {
	case "transaction":
		return OperationKindTransaction
	case "origination":
		return OperationKindOrigination
	case "origination_new":
		return OperationKindOriginationNew
	case "delegation":
		return OperationKindDelegation
	case "register_global_constant":
		return OperationKindRegisterGlobalConstant
	default:
		return 0
	}
}

// String -
func (kind OperationKind) String() string {
	switch kind {
	case OperationKindTransaction:
		return "transaction"
	case OperationKindOrigination:
		return "origination"
	case OperationKindOriginationNew:
		return "origination_new"
	case OperationKindDelegation:
		return "delegation"
	case OperationKindRegisterGlobalConstant:
		return "register_global_constant"
	default:
		return ""
	}
}

// Scan -
func (kind *OperationKind) Scan(value interface{}) error {
	*kind = OperationKind(value.(int64))
	return nil
}

// Value -
func (kind OperationKind) Value() (driver.Value, error) { return int(kind), nil }

const (
	OperationKindTransaction OperationKind = iota + 1
	OperationKindOrigination
	OperationKindOriginationNew
	OperationKindDelegation
	OperationKindRegisterGlobalConstant
)
