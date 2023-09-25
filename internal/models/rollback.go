package models

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
)

type LastAction struct {
	AccountId int64     `bun:"address"`
	Time      time.Time `bun:"time"`
}

//go:generate mockgen -source=$GOFILE -destination=mock/rollback.go -package=mock -typed
type Rollback interface {
	DeleteAll(ctx context.Context, model any, level int64) error
	StatesChangedAtLevel(ctx context.Context, level int64) ([]bigmapdiff.BigMapState, error)
	DeleteBigMapState(ctx context.Context, state bigmapdiff.BigMapState) error
	LastDiff(ctx context.Context, ptr int64, keyHash string, skipRemoved bool) (bigmapdiff.BigMapDiff, error)
	SaveBigMapState(ctx context.Context, state bigmapdiff.BigMapState) error
	GetOperations(ctx context.Context, level int64) ([]operation.Operation, error)
	GetContractsLastAction(ctx context.Context, addressIds ...int64) ([]LastAction, error)
	UpdateContractStats(ctx context.Context, accountId int64, lastAction time.Time, txCount int64) error
	GlobalConstants(ctx context.Context, level int64) ([]contract.GlobalConstant, error)
	Scripts(ctx context.Context, level int64) ([]contract.Script, error)
	DeleteScriptsConstants(ctx context.Context, scriptIds []int64, constantsIds []int64) error

	Commit() error
	Rollback() error
}
