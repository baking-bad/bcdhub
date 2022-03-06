package transfer

import (
	"github.com/baking-bad/bcdhub/internal/events/contracts"
	"github.com/baking-bad/bcdhub/internal/noderpc"
)

// ContractHandlers -
type ContractHandlers map[string]contracts.Contract

// NewContractHandlers -
func NewContractHandlers(rpc noderpc.INode) (ContractHandlers, error) {
	ch := make(ContractHandlers)

	tzbtc, err := contracts.NewTzBTC(rpc)
	if err != nil {
		return nil, err
	}
	ch[tzbtc.Address()] = tzbtc

	kusd, err := contracts.NewKUSD(rpc)
	if err != nil {
		return nil, err
	}
	ch[kusd.Address()] = kusd

	ethtz, err := contracts.NewETHtz(rpc)
	if err != nil {
		return nil, err
	}
	ch[ethtz.Address()] = ethtz

	lbToken, err := contracts.NewLBToken(rpc)
	if err != nil {
		return nil, err
	}
	ch[lbToken.Address()] = lbToken

	usdtz, err := contracts.NewUSDtz(rpc)
	if err != nil {
		return nil, err
	}
	ch[usdtz.Address()] = usdtz

	minter, err := contracts.NewMinter(rpc)
	if err != nil {
		return nil, err
	}
	ch[minter.Address()] = minter

	return ch, nil
}
