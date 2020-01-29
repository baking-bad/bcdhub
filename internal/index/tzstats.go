package index

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/helpers"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/tzstats"
)

// TzStats -
type TzStats struct {
	api *tzstats.TzStats
}

// NewTzStats -
func NewTzStats(host string) *TzStats {
	return &TzStats{
		api: tzstats.NewTzStats(host),
	}
}

type tzStatsHead struct {
	Level     int64  `tzstats:"height"`
	Hash      string `tzstats:"hash"`
	Timestamp int64  `tzstats:"time"`
}

// Name -
func (h tzStatsHead) Name() string {
	return tzstats.TableBlock
}

// GetHead -
func (t *TzStats) GetHead() (Head, error) {
	resp := Head{}

	var head []tzStatsHead
	if err := t.api.Model(tzStatsHead{}).Order("desc").Limit(1).Query(&head); err != nil {
		return resp, err
	}
	if len(head) != 1 {
		return resp, fmt.Errorf("Invalid head response")
	}

	resp.Level = head[0].Level
	resp.Hash = head[0].Hash
	resp.Timestamp = time.Unix(head[0].Timestamp/1000, 0)
	return resp, nil
}

type tzStatsContract struct {
	ID        int64   `tzstats:"row_id"`
	Level     int64   `tzstats:"first_seen"`
	Timestamp int64   `tzstats:"first_seen_time"`
	Balance   float64 `tzstats:"spendable_balance"`
	Manager   string  `tzstats:"manager"`
	Address   string  `tzstats:"address"`
	Delegate  string  `tzstats:"delegate"`
}

// Name -
func (h tzStatsContract) Name() string {
	return tzstats.TableAccount
}

// GetContracts -
func (t *TzStats) GetContracts(startLevel int64) ([]Contract, error) {
	all := make([]Contract, 0)
	rowID := int64(0)

	for {
		var contracts []tzStatsContract
		query := t.api.Model(tzStatsContract{}).Is("is_contract", "1").GreaterThan("first_seen", int(startLevel))
		if rowID > 0 {
			query = query.Is("cursor", fmt.Sprintf("%d", rowID))
		}

		if err := query.Query(&contracts); err != nil {
			return nil, err
		}
		if len(contracts) == 0 {
			return all, nil
		}
		for _, c := range contracts {
			all = append(all, Contract{
				Level:     c.Level,
				Timestamp: time.Unix(c.Timestamp/1000, 0),
				Balance:   int64(c.Balance * 1000000),
				Manager:   c.Manager,
				Address:   c.Address,
				Delegate:  c.Delegate,
			})
			if c.ID > rowID {
				rowID = c.ID
			}
		}

	}
}

type contractOperation struct {
	ID       int64  `tzstats:"row_id"`
	Level    int64  `tzstats:"height"`
	Sender   string `tzstats:"sender"`
	Receiver string `tzstats:"receiver"`
}

// Name -
func (h contractOperation) Name() string {
	return tzstats.TableOperation
}

// GetContractOperationBlocks -
func (t *TzStats) GetContractOperationBlocks(startBlock int, knownContracts []models.Contract) ([]int64, error) {
	all := make(map[int64]struct{})

	addresses := make([]string, len(knownContracts))
	for i := range knownContracts {
		addresses[i] = knownContracts[i].Address
	}

	log.Println("Searching blocks with operations with params...")
	if err := t.getContractOperaionsBlockWithParameters(startBlock, addresses, all); err != nil {
		return nil, err
	}

	log.Println("Searching originations...")
	if err := t.getContractOriginations(startBlock, addresses, all); err != nil {
		return nil, err
	}

	log.Println("Searching spendable contract transactions...")
	spendable := make([]string, 0)
	for _, c := range knownContracts {
		for i := range c.Tags {
			if c.Tags[i] == "spendable" {
				spendable = append(spendable, c.Address)
				break
			}
		}
	}
	if err := t.getSpendableContractOperaions(startBlock, spendable, all); err != nil {
		return nil, err
	}

	buf := make([]int, 0)
	for k := range all {
		buf = append(buf, int(k))
	}

	sort.Ints(buf)

	resp := make([]int64, len(buf))
	for i := range buf {
		resp[i] = int64(buf[i])
	}
	log.Printf("Found %d blocks", len(resp))

	return resp, nil
}

func (t *TzStats) getContractOperaionsBlockWithParameters(startBlock int, knownContracts []string, all map[int64]struct{}) error {
	rowID := int64(0)

	for {
		var operations []contractOperation
		query := t.api.Model(contractOperation{}).NotEquals("parameters", "").GreaterThan("height", startBlock).Limit(50000)
		if rowID > 0 {
			query = query.Is("cursor", fmt.Sprintf("%d", rowID))
		}

		if err := query.Query(&operations); err != nil {
			return err
		}

		if len(operations) == 0 {
			return nil
		}

		for _, op := range operations {
			if _, ok := all[op.Level]; !ok {
				if helpers.StringInArray(op.Sender, knownContracts) || helpers.StringInArray(op.Receiver, knownContracts) {
					all[op.Level] = struct{}{}
				}
			}
			if rowID < op.ID {
				rowID = op.ID
			}
		}
	}
}

func (t *TzStats) getContractOriginations(startBlock int, knownContracts []string, all map[int64]struct{}) error {
	rowID := int64(0)

	for {
		var operations []contractOperation
		query := t.api.Model(contractOperation{}).Is("type", "origination").GreaterThan("height", startBlock).Limit(50000)
		if rowID > 0 {
			query = query.Is("cursor", fmt.Sprintf("%d", rowID))
		}

		if err := query.Query(&operations); err != nil {
			return err
		}

		if len(operations) == 0 {
			return nil
		}

		for _, op := range operations {
			if _, ok := all[op.Level]; !ok {
				if helpers.StringInArray(op.Receiver, knownContracts) {
					all[op.Level] = struct{}{}
				}
			}
			if rowID < op.ID {
				rowID = op.ID
			}
		}
	}
}

func (t *TzStats) getSpendableContractOperaions(startBlock int, spendable []string, all map[int64]struct{}) error {
	rowID := int64(0)

	for {
		var operations []contractOperation
		query := t.api.Model(contractOperation{}).Is("type", "transaction").In("sender", spendable).GreaterThan("height", startBlock).Limit(50000)
		if rowID > 0 {
			query = query.Is("cursor", fmt.Sprintf("%d", rowID))
		}

		if err := query.Query(&operations); err != nil {
			return err
		}

		if len(operations) == 0 {
			return nil
		}

		for _, op := range operations {
			if _, ok := all[op.Level]; !ok {
				all[op.Level] = struct{}{}
			}
			if rowID < op.ID {
				rowID = op.ID
			}
		}
	}
}
