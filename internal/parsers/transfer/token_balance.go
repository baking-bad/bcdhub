package transfer

import (
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
)

// UpdateTokenBalances -
func UpdateTokenBalances(repo tokenbalance.Repository, transfers []*transfer.Transfer) error {
	exists := make(map[string]*tokenbalance.TokenBalance)
	updates := make([]*tokenbalance.TokenBalance, 0)
	for i := range transfers {
		if transfers[i].Status != consts.Applied {
			continue
		}
		idFrom := transfers[i].GetFromTokenBalanceID()
		if idFrom != "" {
			if update, ok := exists[idFrom]; ok {
				update.Value.Sub(update.Value, transfers[i].AmountBigInt)
			} else {
				upd := transfers[i].MakeTokenBalanceUpdate(true, false)
				updates = append(updates, upd)
				exists[idFrom] = upd
			}
		}
		idTo := transfers[i].GetToTokenBalanceID()
		if idTo != "" {
			if update, ok := exists[idTo]; ok {
				update.Value.Add(update.Value, transfers[i].AmountBigInt)
			} else {
				upd := transfers[i].MakeTokenBalanceUpdate(false, false)
				updates = append(updates, upd)
				exists[idTo] = upd
			}
		}
	}

	return repo.Update(updates)
}
