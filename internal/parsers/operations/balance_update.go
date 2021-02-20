package operations

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/balanceupdate"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/tidwall/gjson"
)

// BalanceUpdate -
type BalanceUpdate struct {
	operation operation.Operation
	root      string
}

// NewBalanceUpdate -
func NewBalanceUpdate(root string, operation operation.Operation) BalanceUpdate {
	return BalanceUpdate{operation, root}
}

// Parse -
func (b BalanceUpdate) Parse(data gjson.Result) []*balanceupdate.BalanceUpdate {
	if b.root != "" {
		b.root = fmt.Sprintf("%s.", b.root)
	}
	filter := fmt.Sprintf(`%sbalance_updates.#(kind="contract")#`, b.root)

	contracts := data.Get(filter).Array()
	bu := make([]*balanceupdate.BalanceUpdate, 0)
	for i := range contracts {
		address := contracts[i].Get("contract").String()
		if !bcd.IsContract(address) {
			continue
		}
		bu = append(bu, &balanceupdate.BalanceUpdate{
			ID:            helpers.GenerateID(),
			Change:        contracts[i].Get("change").Int(),
			Network:       b.operation.Network,
			Contract:      address,
			OperationHash: b.operation.Hash,
			ContentIndex:  b.operation.ContentIndex,
			Nonce:         b.operation.Nonce,
			Level:         b.operation.Level,
		})
	}
	return bu
}
