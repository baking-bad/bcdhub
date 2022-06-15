package types

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
	case "tx_rollup_origination":
		return OperationKindTxRollupOrigination
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
	case OperationKindTxRollupOrigination:
		return "tx_rollup_origination"
	default:
		return ""
	}
}

const (
	OperationKindTransaction OperationKind = iota + 1
	OperationKindOrigination
	OperationKindOriginationNew
	OperationKindDelegation
	OperationKindRegisterGlobalConstant
	OperationKindTxRollupOrigination
)
