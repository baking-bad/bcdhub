package transfer

import "github.com/baking-bad/bcdhub/internal/models/tzip"

// Repository -
type Repository interface {
	Get(ctx GetContext) (Pageable, error)
	GetAll(network string, level int64) ([]Transfer, error)
	GetBalances(string, string, int64, ...TokenBalance) (map[TokenBalance]int64, error)
	GetTokenSupply(network, address string, tokenID int64) (result TokenSupply, err error)
	GetToken24HoursVolume(network, contract string, initiators, entrypoints []string, tokenID int64) (float64, error)
	GetTokenVolumeSeries(network, period string, contracts []string, entrypoints []tzip.DAppContract, tokenID uint) ([][]int64, error)
}
