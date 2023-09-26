package models

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	smartrollup "github.com/baking-bad/bcdhub/internal/models/smart_rollup"
	"github.com/baking-bad/bcdhub/internal/models/stats"
	"github.com/baking-bad/bcdhub/internal/models/ticket"
)

//go:generate mockgen -source=$GOFILE -destination=mock/general.go -package=mock -typed
type GeneralRepository interface {
	InitDatabase(ctx context.Context) error
	TablesExist(ctx context.Context) bool
	CreateIndex(ctx context.Context, name, columns string, model any) error
	IsRecordNotFound(err error) bool

	// Drop - drops full database
	Drop(ctx context.Context) error
}

//go:generate mockgen -source=$GOFILE -destination=mock/general.go -package=mock -typed
type Transaction interface {
	Save(ctx context.Context, data any) error
	Migrations(ctx context.Context, migrations ...*migration.Migration) error
	GlobalConstants(ctx context.Context, constants ...*contract.GlobalConstant) error
	BigMapStates(ctx context.Context, states ...*bigmapdiff.BigMapState) error
	BigMapDiffs(ctx context.Context, bigmapdiffs ...*bigmapdiff.BigMapDiff) error
	BigMapActions(ctx context.Context, bigmapdiffs ...*bigmapaction.BigMapAction) error
	Accounts(ctx context.Context, accounts ...*account.Account) error
	SmartRollups(ctx context.Context, rollups ...*smartrollup.SmartRollup) error
	Operations(ctx context.Context, operations ...*operation.Operation) error
	TickerUpdates(ctx context.Context, updates ...*ticket.TicketUpdate) error
	Contracts(ctx context.Context, contracts ...*contract.Contract) error
	Scripts(ctx context.Context, scripts ...*contract.Script) error
	ScriptConstant(ctx context.Context, data ...*contract.ScriptConstants) error
	Block(ctx context.Context, block *block.Block) error
	Protocol(ctx context.Context, proto *protocol.Protocol) error
	UpdateStats(ctx context.Context, stats stats.Stats) error

	ToBabylon(ctx context.Context) error
	BabylonUpdateNonDelegator(ctx context.Context, contract *contract.Contract) error
	ToJakarta(ctx context.Context) error
	JakartaVesting(ctx context.Context, contract *contract.Contract) error
	JakartaUpdateNonDelegator(ctx context.Context, contract *contract.Contract) error
	BabylonBigMapStates(ctx context.Context, state *bigmapdiff.BigMapState) error

	Commit() error
	Rollback() error
}
