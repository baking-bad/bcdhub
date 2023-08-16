package types

import "github.com/baking-bad/bcdhub/internal/bcd/consts"

// OperationKind -
type OperationKind int

// NewOperationKind -
func NewOperationKind(value string) OperationKind {
	switch value {
	case consts.Transaction:
		return OperationKindTransaction
	case consts.Origination:
		return OperationKindOrigination
	case consts.OriginationNew:
		return OperationKindOriginationNew
	case consts.Delegation:
		return OperationKindDelegation
	case consts.RegisterGlobalConstant:
		return OperationKindRegisterGlobalConstant
	case consts.TxRollupOrigination:
		return OperationKindTxRollupOrigination
	case consts.Event:
		return OperationKindEvent
	case consts.TransferTicket:
		return OperationKindTransferTicket
	case consts.SrOriginate:
		return OperationKindSrOrigination
	case consts.SrExecuteOutboxMessage:
		return OperationKindSrExecuteOutboxMessage
	default:
		return 0
	}
}

// String -
func (kind OperationKind) String() string {
	switch kind {
	case OperationKindTransaction:
		return consts.Transaction
	case OperationKindOrigination:
		return consts.Origination
	case OperationKindOriginationNew:
		return consts.OriginationNew
	case OperationKindDelegation:
		return consts.Delegation
	case OperationKindRegisterGlobalConstant:
		return consts.RegisterGlobalConstant
	case OperationKindTxRollupOrigination:
		return consts.TxRollupOrigination
	case OperationKindEvent:
		return consts.Event
	case OperationKindTransferTicket:
		return consts.TransferTicket
	case OperationKindSrOrigination:
		return consts.SrOriginate
	case OperationKindSrExecuteOutboxMessage:
		return consts.SrExecuteOutboxMessage
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
	OperationKindEvent
	OperationKindTransferTicket
	OperationKindSrOrigination
	OperationKindSrExecuteOutboxMessage
)
