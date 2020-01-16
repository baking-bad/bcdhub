package index

import (
	"github.com/aopoltorzhicky/bcdhub/internal/tzkt"
	"strings"
	"time"
)

// TzKT -
type TzKT struct {
	api *tzkt.TzKT
}

// NewTzKT -
func NewTzKT(host string, timeout time.Duration) *TzKT {
	return &TzKT{
		api: tzkt.NewTzKT(host, timeout),
	}
}

// GetHead -
func (t *TzKT) GetHead() (Head, error) {
	resp := Head{}
	head, err := t.api.GetHead()
	if err != nil {
		return resp, err
	}
	resp.Level = head.Level
	resp.Hash = head.Hash
	resp.Timestamp = head.Timestamp
	return resp, err
}

// GetContracts -
func (t *TzKT) GetContracts(startLevel int64) ([]Contract, error) {
	resp := make([]Contract, 0)

	count, err := t.api.GetOriginationsCount()
	if err != nil {
		return nil, err
	}

	limit := int64(1000)
	for ; count > 0; count = count - limit {
		page := int64(count / limit)
		ops, err := t.api.GetOriginations(page, limit)
		if err != nil {
			return resp, err
		}

		for i := len(ops) - 1; i >= 0; i-- {
			op := ops[i]
			if op.Level < startLevel {
				count = 0
				break
			}
			if op.Status == "applied" && op.OriginatedContract.Kind != "delegator_contract" {
				resp = append(resp, Contract{
					Level:     op.Level,
					Timestamp: op.Timestamp,
					Counter:   op.Counter,
					Balance:   op.ContractBalance,
					Manager:   op.ContractManager.Address,
					Delegate:  op.ContractDelegate.Address,
					Address:   op.OriginatedContract.Address,
				})
			}
		}
	}

	if startLevel == 0 {
		count, err = t.api.GetSystemOperationsCount()
		if err != nil {
			return nil, err
		}

		page := int64(0)
		for ; count > 0; count = count - limit {
			ops, err := t.api.GetSystemOperations(page, limit)
			if err != nil {
				return resp, err
			}

			for _, op := range ops {
				if op.Level > 1 {
					count = 0
					break
				}
				if op.Kind == "bootstrap" && strings.HasPrefix(op.Account.Address, "KT1") {
					resp = append(resp, Contract{
						Level:     op.Level,
						Timestamp: op.Timestamp,
						Address:   op.Account.Address,
					})
				}
			}

			page = page + 1
		}
	}

	return resp, nil
}

// GetContractOperationBlocks -
func (t *TzKT) GetContractOperationBlocks(startBlock int) ([]int64, error) {
	return nil, nil
}
