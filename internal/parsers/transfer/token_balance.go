package transfer

import (
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
)

// UpdateTokenBalances -
func UpdateTokenBalances(transfers []*transfer.Transfer) []*tokenbalance.TokenBalance {
	exists := make(map[string]*tokenbalance.TokenBalance)
	for i := range transfers {
		if transfers[i].Status != consts.Applied {
			continue
		}
		idFrom := transfers[i].GetFromTokenBalanceID()
		if idFrom != "" {
			if update, ok := exists[idFrom]; ok {
				update.Value.Sub(update.Value, transfers[i].Value)
			} else {
				exists[idFrom] = transfers[i].MakeTokenBalanceUpdate(true, false)
			}
		}
		idTo := transfers[i].GetToTokenBalanceID()
		if idTo != "" {
			if update, ok := exists[idTo]; ok {
				update.Value.Add(update.Value, transfers[i].Value)
			} else {
				exists[idTo] = transfers[i].MakeTokenBalanceUpdate(false, false)
			}
		}
	}

	updates := make([]*tokenbalance.TokenBalance, 0, len(exists))
	for _, upd := range exists {
		updates = append(updates, upd)
	}

	return updates
}
