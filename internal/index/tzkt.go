package index

import (
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/tzkt"
)

// TzKT -
type TzKT struct {
	api *tzkt.TzKT

	lastContractsPage int64
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

	end := false
	for !end {
		contracts, err := t.api.GetAccounts(tzkt.ContractKindAll, t.lastContractsPage, 1000)
		if err != nil {
			return nil, err
		}
		for _, contract := range contracts {
			if contract.FirstActivity <= startLevel {
				continue
			}

			data := Contract{
				Address:   contract.Address,
				Level:     contract.FirstActivity,
				Timestamp: contract.FirstActivityTime,
				Balance:   contract.Balance,
			}
			if contract.Manager != nil {
				data.Manager = contract.Manager.Address
			}
			if contract.Delegate != nil {
				data.Delegate = contract.Delegate.Address
			}
			resp = append(resp, data)

		}
		if len(contracts) == 1000 {
			t.lastContractsPage++
		} else {
			end = true
		}
	}

	return resp, nil
}

// GetContractOperationBlocks -
func (t *TzKT) GetContractOperationBlocks(startBlock, endBlock int, knownContracts map[string]struct{}, spendable map[string]struct{}) ([]int64, error) {
	resp := make([]int64, endBlock-startBlock+1)
	for i := startBlock; i < endBlock+1; i++ {
		resp[i] = int64(i)
	}
	return resp, nil
}
