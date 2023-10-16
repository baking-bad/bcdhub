package smartrollup

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/uptrace/bun"
)

// SmartRollup - entity for smart rollup
type SmartRollup struct {
	bun.BaseModel `bun:"smart_rollup"`

	ID        int64 `bun:"id,pk,notnull,autoincrement"`
	Level     int64
	Timestamp time.Time

	Size      uint64
	AddressId int64
	Address   account.Account `bun:",rel:belongs-to"`

	GenesisCommitmentHash string `bun:"genesis_commitment_hash,type:text"`
	PvmKind               string `bun:"pvm_kind,type:text"`
	Kernel                []byte `bun:",type:bytea"`
	Type                  []byte `bun:",type:bytea"`
}

// GetID -
func (sr *SmartRollup) GetID() int64 {
	return sr.ID
}

func (SmartRollup) TableName() string {
	return "smart_rollup"
}

// LogFields -
func (sr *SmartRollup) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"address": sr.Address.Address,
		"block":   sr.Level,
	}
}
