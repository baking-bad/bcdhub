package contracts

import (
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
)

const (
	mint       = "mint"
	burn       = "burn"
	mintOrBurn = "mintOrBurn"
)

// Contract -
type Contract interface {
	Address() string
	HasHandler(entrypoint string) bool
	Handler(parameters *types.Parameters) ([]tokenbalance.TokenBalance, error)
}
