package store

import (
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

// AddOperations -
func (store *Store) AddOperations(operations ...*operation.Operation) {
	store.Operations = append(store.Operations, operations...)

	store.Stats.OperationsCount += len(operations)
	for i := range operations {
		switch operations[i].Kind {
		case types.OperationKindEvent:
			store.Stats.EventsCount += 1
		case types.OperationKindOrigination, types.OperationKindOriginationNew:
			store.Stats.OriginationsCount += 1
		case types.OperationKindSrOrigination:
			store.Stats.SrOriginationsCount += 1
		case types.OperationKindTransaction:
			store.Stats.TransactionsCount += 1
		case types.OperationKindRegisterGlobalConstant:
			store.Stats.RegisterGlobalConstantCount += 1
		case types.OperationKindSrExecuteOutboxMessage:
			store.Stats.SrExecutesCount += 1
		}
	}
}
