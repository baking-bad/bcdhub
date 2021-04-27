package transfer

import "github.com/baking-bad/bcdhub/internal/models/tzip"

// Repository -
type Repository interface {
	Get(ctx GetContext) (Pageable, error)
	GetAll(network string, level int64) ([]Transfer, error)
	GetTransfered(network, address string, tokenID int64) (result uint64, err error)
	GetToken24HoursVolume(network, contract string, initiators, entrypoints []string, tokenID int64) (float64, error)
	GetTokenVolumeSeries(network, period string, contracts []string, entrypoints []tzip.DAppContract, tokenID uint) ([][]float64, error)
}
