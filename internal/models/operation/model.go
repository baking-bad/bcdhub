package operation

import (
	"encoding/hex"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/ticket"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/uptrace/bun"
)

// Operation -
type Operation struct {
	bun.BaseModel `bun:"operations"`

	ID                                 int64 `bun:"id,pk,notnull,autoincrement"`
	ContentIndex                       int64
	Level                              int64
	Counter                            int64
	Fee                                int64
	GasLimit                           int64
	StorageLimit                       int64
	Amount                             int64
	ConsumedGas                        int64
	StorageSize                        int64
	PaidStorageSizeDiff                int64
	Burned                             int64
	AllocatedDestinationContractBurned int64
	ProtocolID                         int64 `bun:"protocol_id,type:SMALLINT"`
	TicketUpdatesCount                 int
	BigMapDiffsCount                   int
	Tags                               types.Tags
	Nonce                              *int64 `bun:"nonce,nullzero"`

	InitiatorID   int64
	Initiator     account.Account `bun:"rel:belongs-to"`
	SourceID      int64
	Source        account.Account `bun:"rel:belongs-to"`
	DestinationID int64
	Destination   account.Account `bun:"rel:belongs-to"`
	DelegateID    int64
	Delegate      account.Account `bun:"rel:belongs-to"`

	Timestamp time.Time             `bun:"timestamp,pk,notnull"`
	Status    types.OperationStatus `bun:"status,type:SMALLINT"`
	Kind      types.OperationKind   `bun:"kind,type:SMALLINT"`

	Entrypoint      types.NullString `bun:"entrypoint,type:text"`
	Tag             types.NullString `bun:"tag,type:text"`
	Hash            []byte
	Parameters      []byte
	DeffatedStorage []byte
	Payload         []byte
	PayloadType     []byte
	Script          []byte `bun:"-"`

	Errors tezerrors.Errors `bun:"errors,type:bytea"`

	AST *ast.Script `bun:"-"`

	BigMapDiffs   []*bigmapdiff.BigMapDiff     `bun:"rel:has-many"`
	BigMapActions []*bigmapaction.BigMapAction `bun:"rel:has-many"`
	TicketUpdates []*ticket.TicketUpdate       `bun:"rel:has-many"`

	AllocatedDestinationContract bool
	Internal                     bool
}

// GetID -
func (o *Operation) GetID() int64 {
	return o.ID
}

func (o *Operation) TableName() string {
	return "operations"
}

// LogFields -
func (o *Operation) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"hash":  hex.EncodeToString(o.Hash),
		"block": o.Level,
	}
}

// SetAllocationBurn -
func (o *Operation) SetAllocationBurn(constants protocol.Constants) {
	o.AllocatedDestinationContractBurned = 257 * constants.CostPerByte
}

// SetBurned -
func (o *Operation) SetBurned(constants protocol.Constants) {
	if o.Status != types.OperationStatusApplied {
		return
	}
	var burned int64

	if o.PaidStorageSizeDiff != 0 {
		burned += o.PaidStorageSizeDiff * constants.CostPerByte
	}

	if o.AllocatedDestinationContract {
		o.SetAllocationBurn(constants)
		burned += o.AllocatedDestinationContractBurned
	}

	o.Burned = burned
}

// IsEntrypoint -
func (o *Operation) IsEntrypoint(entrypoint string) bool {
	return o.Entrypoint.EqualString(entrypoint)
}

// IsOrigination -
func (o *Operation) IsOrigination() bool {
	return o.Kind == types.OperationKindOrigination || o.Kind == types.OperationKindOriginationNew
}

// IsTransaction -
func (o *Operation) IsTransaction() bool {
	return o.Kind == types.OperationKindTransaction
}

// IsImplicit  -
func (o *Operation) IsImplicit() bool {
	return len(o.Hash) == 0
}

// IsApplied -
func (o *Operation) IsApplied() bool {
	return o.Status == types.OperationStatusApplied
}

// IsCall -
func (o *Operation) IsCall() bool {
	return (bcd.IsContract(o.Destination.Address) || bcd.IsSmartRollupHash(o.Destination.Address)) && len(o.Parameters) > 0
}

func (o *Operation) CanHasStorageDiff() bool {
	return o.IsApplied() &&
		len(o.DeffatedStorage) > 0 &&
		(o.IsCall() || o.IsOrigination() || o.IsImplicit())
}

// Result -
type Result struct {
	Status                       string
	ConsumedGas                  int64
	StorageSize                  int64
	PaidStorageSizeDiff          int64
	AllocatedDestinationContract bool
	Originated                   string
	Errors                       []*tezerrors.Error
}

// Stats -
type Stats struct {
	Count      int64
	LastAction time.Time
}

// Pageable -
type Pageable struct {
	Operations []Operation
	LastID     string
}

// TokenMethodUsageStats -
type TokenMethodUsageStats struct {
	Count       int64
	ConsumedGas int64
}

// TokenUsageStats -
type TokenUsageStats map[string]TokenMethodUsageStats
