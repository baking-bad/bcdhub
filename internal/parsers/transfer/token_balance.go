package transfer

import (
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

// UpdateTokenBalances -
func UpdateTokenBalances(transfers []*transfer.Transfer) []*tokenbalance.TokenBalance {
	exists := make(map[string]*tokenbalance.TokenBalance)
	updates := make([]*tokenbalance.TokenBalance, 0)
	for i := range transfers {
		if transfers[i].Status != types.OperationStatusApplied {
			continue
		}
		idFrom := transfers[i].GetFromTokenBalanceID()
		if idFrom != "" {
			if update, ok := exists[idFrom]; ok {
				update.Balance = update.Balance.Sub(transfers[i].Amount)
			} else {
				exists[idFrom] = transfers[i].MakeTokenBalanceUpdate(true, false)
				updates = append(updates, exists[idFrom])
			}
		}
		idTo := transfers[i].GetToTokenBalanceID()
		if idTo != "" {
			if update, ok := exists[idTo]; ok {
				update.Balance = update.Balance.Add(transfers[i].Amount)
			} else {
				exists[idTo] = transfers[i].MakeTokenBalanceUpdate(false, false)
				updates = append(updates, exists[idTo])
			}
		}
	}
	return updates
}
