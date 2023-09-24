package operation

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models/account"
)

//go:generate mockgen -source=$GOFILE -destination=../mock/operation/mock.go -package=operation -typed
type Repository interface {
	GetByAccount(ctx context.Context, acc account.Account, size uint64, filters map[string]interface{}) (Pageable, error)
	// Last -  get last operation by `filters` with not empty deffated_storage.
	Last(ctx context.Context, filter map[string]interface{}, lastID int64) (Operation, error)
	GetByHash(ctx context.Context, hash []byte) ([]Operation, error)
	GetByHashAndCounter(ctx context.Context, hash []byte, counter int64) ([]Operation, error)
	GetImplicitOperation(ctx context.Context, counter int64) (Operation, error)
	OPG(ctx context.Context, address string, size, lastID int64) ([]OPG, error)
	Origination(ctx context.Context, accountID int64) (Operation, error)

	// GetOperations - get operation by `filter`. `Size` - if 0 - return all, else certain `size` operations.
	// `Sort` - sort by time and content index by desc
	Get(ctx context.Context, filter map[string]interface{}, size int64, sort bool) ([]Operation, error)
	GetByID(ctx context.Context, id int64) (Operation, error)

	ListEvents(ctx context.Context, accountID int64, size, offset int64) ([]Operation, error)
	EventsCount(ctx context.Context, accountID int64) (int, error)
	ContractStats(ctx context.Context, address string) (ContractStats, error)
}
