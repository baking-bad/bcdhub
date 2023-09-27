package models

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/stats"
)

type LastAction struct {
	AccountId int64     `bun:"address"`
	Time      time.Time `bun:"time"`
}

//go:generate mockgen -source=$GOFILE -destination=mock/rollback.go -package=mock -typed
type Rollback interface {
	DeleteAll(ctx context.Context, model any, level int64) (int, error)
	StatesChangedAtLevel(ctx context.Context, level int64) ([]bigmapdiff.BigMapState, error)
	DeleteBigMapState(ctx context.Context, state bigmapdiff.BigMapState) error
	LastDiff(ctx context.Context, ptr int64, keyHash string, skipRemoved bool) (bigmapdiff.BigMapDiff, error)
	SaveBigMapState(ctx context.Context, state bigmapdiff.BigMapState) error
	GetOperations(ctx context.Context, level int64) ([]operation.Operation, error)
	GetMigrations(ctx context.Context, level int64) ([]migration.Migration, error)
	GetLastAction(ctx context.Context, addressIds ...int64) ([]LastAction, error)
	UpdateAccountStats(ctx context.Context, account account.Account) error
	GlobalConstants(ctx context.Context, level int64) ([]contract.GlobalConstant, error)
	Scripts(ctx context.Context, level int64) ([]contract.Script, error)
	DeleteScriptsConstants(ctx context.Context, scriptIds []int64, constantsIds []int64) error
	Protocols(ctx context.Context, level int64) error
	UpdateStats(ctx context.Context, stats stats.Stats) error

	Commit() error
	Rollback() error
}
