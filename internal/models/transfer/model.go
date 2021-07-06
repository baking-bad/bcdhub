package transfer

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Transfer -
type Transfer struct {
	ID          int64                 `json:"-"`
	Network     types.Network         `json:"network" gorm:"type:SMALLINT;index:transfers_network_idx;index:transfers_token_idx"`
	Contract    string                `json:"contract" gorm:"index:transfers_token_idx"`
	Initiator   string                `json:"initiator"`
	Status      types.OperationStatus `json:"status" gorm:"type:SMALLINT"`
	Timestamp   time.Time             `json:"timestamp" gorm:"index:transfers_timestamp_idx"`
	Level       int64                 `json:"level" gorm:"index:transfers_network_idx;index:transfers_level_idx"`
	From        string                `json:"from" gorm:"index:transfers_from_idx"`
	To          string                `json:"to" gorm:"index:transfers_to_idx"`
	TokenID     uint64                `json:"token_id" gorm:"type:numeric(50,0);index:transfers_token_idx"`
	Amount      decimal.Decimal       `json:"amount" gorm:"type:numeric(100,0)"`
	Parent      string                `json:"parent,omitempty"`
	Entrypoint  string                `json:"entrypoint,omitempty"`
	OperationID int64                 `json:"-"`
}

// GetID -
func (t *Transfer) GetID() int64 {
	return t.ID
}

// GetIndex -
func (t *Transfer) GetIndex() string {
	return "transfers"
}

// Save -
func (t *Transfer) Save(tx *gorm.DB) error {
	return tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Save(t).Error
}

// LogFields -
func (t *Transfer) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"network":  t.Network.String(),
		"contract": t.Contract,
		"block":    t.Level,
		"from":     t.From,
		"to":       t.To,
	}
}

// GetFromTokenBalanceID -
func (t *Transfer) GetFromTokenBalanceID() string {
	if t.From != "" {
		return fmt.Sprintf("%s_%s_%s_%d", t.Network, t.From, t.Contract, t.TokenID)
	}
	return ""
}

// GetToTokenBalanceID -
func (t *Transfer) GetToTokenBalanceID() string {
	if t.To != "" {
		return fmt.Sprintf("%s_%s_%s_%d", t.Network, t.To, t.Contract, t.TokenID)
	}
	return ""
}

// MakeTokenBalanceUpdate -
func (t *Transfer) MakeTokenBalanceUpdate(from, rollback bool) *tokenbalance.TokenBalance {
	tb := &tokenbalance.TokenBalance{
		Network:  t.Network,
		Contract: t.Contract,
		TokenID:  t.TokenID,
		Balance:  decimal.Zero,
	}
	switch {
	case from && rollback:
		tb.Address = t.From
		tb.Balance = t.Amount
	case !from && rollback:
		tb.Address = t.To
		tb.Balance = t.Amount.Neg()
	case from && !rollback:
		tb.Address = t.From
		tb.Balance = t.Amount.Neg()
	case !from && !rollback:
		tb.Address = t.To
		tb.Balance = t.Amount
	}
	return tb
}

// Pageable -
type Pageable struct {
	Transfers []Transfer `json:"transfers"`
	Total     int64      `json:"total"`
	LastID    string     `json:"last_id"`
}

// Balance -
type Balance struct {
	Balance decimal.Decimal
	Address string
	TokenID uint64
}
