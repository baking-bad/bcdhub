package operations

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

// BalanceUpdate -
type BalanceUpdate struct {
	root string
}

// NewBalanceUpdate -
func NewBalanceUpdate(root string) BalanceUpdate {
	return BalanceUpdate{root}
}

// Parse -
func (b BalanceUpdate) Parse(data gjson.Result) []models.BalanceUpdate {
	if b.root != "" {
		b.root = fmt.Sprintf("%s.", b.root)
	}
	filter := fmt.Sprintf("%sbalance_updates.#(kind==\"contract\")#", b.root)

	contracts := data.Get(filter).Array()
	bu := make([]models.BalanceUpdate, len(contracts))
	for i := range contracts {
		bu[i] = models.BalanceUpdate{
			Kind:     contracts[i].Get("kind").String(),
			Contract: contracts[i].Get("contract").String(),
			Change:   contracts[i].Get("change").Int(),
		}
	}
	return bu
}
