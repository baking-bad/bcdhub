package transfer

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/sirupsen/logrus"
)

// Transfer -
type Transfer struct {
	ID          string    `json:"-"`
	IndexedTime int64     `json:"indexed_time"`
	Network     string    `json:"network"`
	Contract    string    `json:"contract"`
	Initiator   string    `json:"initiator"`
	Hash        string    `json:"hash"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
	Level       int64     `json:"level"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	TokenID     int64     `json:"token_id"`
	Amount      float64   `json:"amount"`
	Counter     int64     `json:"counter"`
	Nonce       *int64    `json:"nonce,omitempty"`
	Parent      string    `json:"parent,omitempty"`
}

// GetID -
func (t *Transfer) GetID() string {
	return t.ID
}

// GetIndex -
func (t *Transfer) GetIndex() string {
	return "transfer"
}

// GetQueues -
func (t *Transfer) GetQueues() []string {
	return nil
}

// MarshalToQueue -
func (t *Transfer) MarshalToQueue() ([]byte, error) {
	return nil, nil
}

// LogFields -
func (t *Transfer) LogFields() logrus.Fields {
	return logrus.Fields{
		"network":  t.Network,
		"contract": t.Contract,
		"block":    t.Level,
		"from":     t.From,
		"to":       t.To,
	}
}

// EmptyTransfer -
func EmptyTransfer(o operation.Operation) *Transfer {
	return &Transfer{
		ID:          helpers.GenerateID(),
		IndexedTime: o.IndexedTime,
		Network:     o.Network,
		Contract:    o.Destination,
		Hash:        o.Hash,
		Status:      o.Status,
		Timestamp:   o.Timestamp,
		Level:       o.Level,
		Initiator:   o.Source,
		Counter:     o.Counter,
		Nonce:       o.Nonce,
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
	}
	switch {
	case from && rollback:
		tb.Address = t.From
		tb.Balance = int64(t.Amount)
	case !from && rollback:
		tb.Address = t.To
		tb.Balance = -int64(t.Amount)
	case from && !rollback:
		tb.Address = t.From
		tb.Balance = -int64(t.Amount)
	case !from && !rollback:
		tb.Address = t.To
		tb.Balance = int64(t.Amount)
	}
	return tb
}

// TokenBalance -
type TokenBalance struct {
	Address string
	TokenID int64
}

// TokenSupply -
type TokenSupply struct {
	Supply     float64 `json:"supply"`
	Transfered float64 `json:"transfered"`
}

// Pageable -
type Pageable struct {
	Transfers []Transfer `json:"transfers"`
	Total     int64      `json:"total"`
	LastID    string     `json:"last_id"`
}
