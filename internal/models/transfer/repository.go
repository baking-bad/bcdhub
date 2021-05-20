package transfer

import (
	"github.com/baking-bad/bcdhub/internal/models/dapp"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

// Repository -
type Repository interface {
	Get(ctx GetContext) (Pageable, error)
	GetAll(network types.Network, level int64) ([]Transfer, error)
	GetTransfered(network types.Network, contract string, tokenID uint64) (result float64, err error)
	GetToken24HoursVolume(network types.Network, contract string, initiators, entrypoints []string, tokenID uint64) (float64, error)
	GetTokenVolumeSeries(network types.Network, period string, contracts []string, entrypoints []dapp.DAppContract, tokenID uint64) ([][]float64, error)
	CalcBalances(network types.Network, contract string) ([]Balance, error)
}
