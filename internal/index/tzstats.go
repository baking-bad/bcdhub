package index

import (
	"fmt"
	"time"

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
	first := true
	empty := false
	rowID := int64(0)

	for first || !empty {
		first = false
		var contracts []tzStatsContract
		query := t.api.Model(tzStatsContract{}).Is("is_contract", "1").GreaterThan("first_seen", int(startLevel))
		if rowID > 0 {
			query = query.Is("cursor", fmt.Sprintf("%d", rowID))
		}

		if err := query.Query(&contracts); err != nil {
			return nil, err
		}
		empty = len(contracts) == 0
		if !empty {
			resp := make([]Contract, len(contracts))
			for idx, c := range contracts {
				resp[idx] = Contract{
					Level:     c.Level,
					Timestamp: time.Unix(c.Timestamp/1000, 0),
					Balance:   int64(c.Balance * 1000000),
					Manager:   c.Manager,
					Address:   c.Address,
					Delegate:  c.Delegate,
				}
				if c.ID > rowID {
					rowID = c.ID
				}
			}
			all = append(all, resp...)
		}
	}
	return all, nil
}
