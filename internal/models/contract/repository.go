package contract

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
)

//go:generate mockgen -source=$GOFILE -destination=../mock/contract/mock.go -package=contract -typed
type Repository interface {
	Get(ctx context.Context, address string) (Contract, error)
	RecentlyCalled(ctx context.Context, offset, size int64) ([]Contract, error)
	Count(ctx context.Context) (int, error)
	Script(ctx context.Context, address string, symLink string) (Script, error)

	// ScriptPart - returns part of script type. Part can be `storage`, `parameter` or `code`.
	ScriptPart(ctx context.Context, address string, symLink, part string) ([]byte, error)
	FindOne(ctx context.Context, tags types.Tags) (Contract, error)
	AllExceptDelegators(ctx context.Context) ([]Contract, error)
}

//go:generate mockgen -source=$GOFILE -destination=../mock/contract/mock.go -package=contract -typed
type ScriptRepository interface {
	ByHash(ctx context.Context, hash string) (Script, error)
	Code(ctx context.Context, id int64) ([]byte, error)
	Parameter(ctx context.Context, id int64) ([]byte, error)
	Storage(ctx context.Context, id int64) ([]byte, error)
	Views(ctx context.Context, id int64) ([]byte, error)
}

//go:generate mockgen -source=$GOFILE -destination=../mock/contract/mock.go -package=contract -typed
type ConstantRepository interface {
	Get(ctx context.Context, address string) (GlobalConstant, error)
	All(ctx context.Context, addresses ...string) ([]GlobalConstant, error)
	List(ctx context.Context, size, offset int64, orderBy, sort string) ([]ListGlobalConstantItem, error)
	ForContract(ctx context.Context, address string, size, offset int64) ([]GlobalConstant, error)
	ContractList(ctx context.Context, address string, size, offset int64) ([]Contract, error)
}

// ListGlobalConstantItem -
type ListGlobalConstantItem struct {
	Timestamp  time.Time `json:"timestamp" pg:"timestamp"`
	Level      int64     `json:"level" pg:"level"`
	Address    string    `json:"address" pg:"address"`
	LinksCount uint64    `json:"links_count" pg:"links_count"`
}
