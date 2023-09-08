package operation

import (
	"github.com/baking-bad/bcdhub/internal/models/account"
)

//go:generate mockgen -source=$GOFILE -destination=../mock/operation/mock.go -package=operation -typed
type Repository interface {
	GetByAccount(acc account.Account, size uint64, filters map[string]interface{}) (Pageable, error)
	// Last -  get last operation by `filters` with not empty deffated_storage.
	Last(filter map[string]interface{}, lastID int64) (Operation, error)
	GetByHash(hash []byte) ([]Operation, error)
	GetByHashAndCounter(hash []byte, counter int64) ([]Operation, error)
	GetImplicitOperation(counter int64) (Operation, error)
	OPG(address string, size, lastID int64) ([]OPG, error)
	Origination(accountID int64) (Operation, error)

	// GetOperations - get operation by `filter`. `Size` - if 0 - return all, else certain `size` operations.
	// `Sort` - sort by time and content index by desc
	Get(filter map[string]interface{}, size int64, sort bool) ([]Operation, error)

	GetByIDs(ids ...int64) ([]Operation, error)
	GetByID(id int64) (Operation, error)

	ListEvents(accountID int64, size, offset int64) ([]Operation, error)
	EventsCount(accountID int64) (int, error)
	ContractStats(address string) (ContractStats, error)
}
