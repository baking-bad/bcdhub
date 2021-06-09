package metrics

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/events"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	transferParsers "github.com/baking-bad/bcdhub/internal/parsers/transfer"
)

// CreateTokenMetadata -
func (h *Handler) CreateTokenMetadata(rpc noderpc.INode, sharePath string, c *contract.Contract, ipfs ...string) ([]models.Model, error) {
	result := make([]models.Model, 0)

	transfers, err := h.ExecuteInitialStorageEvent(rpc, c.Network, c.Address)
	if err != nil {
		return nil, err
	}
	for i := range transfers {
		result = append(result, transfers[i])
	}

	return result, nil
}

// ExecuteInitialStorageEvent -
func (h *Handler) ExecuteInitialStorageEvent(rpc noderpc.INode, network types.Network, contract string) ([]models.Model, error) {
	tzip, err := h.TZIP.Get(network, contract)
	if err != nil {
		if h.Storage.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(tzip.Events) == 0 {
		return nil, nil
	}

	ops, err := h.Operations.Get(map[string]interface{}{
		"destination": contract,
		"network":     network,
		"kind":        types.OperationKindOrigination,
	}, 1, false)
	if err != nil {
		if h.Storage.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(ops) != 1 {
		return nil, nil
	}

	origination := ops[0]

	protocol, err := h.Protocol.GetByID(origination.ProtocolID)
	if err != nil {
		return nil, err
	}

	state, err := h.Blocks.Last(network)
	if err != nil {
		return nil, err
	}

	newModels := make([]models.Model, 0)

	for i := range tzip.Events {
		for j := range tzip.Events[i].Implementations {
			impl := tzip.Events[i].Implementations[j]
			if impl.MichelsonInitialStorageEvent != nil && !impl.MichelsonInitialStorageEvent.Empty() {
				event, err := events.NewMichelsonInitialStorage(impl, tzip.Events[i].Name)
				if err != nil {
					return nil, err
				}

				ops, err := rpc.GetOPG(origination.Level)
				if err != nil {
					return nil, err
				}

				var opg noderpc.OperationGroup
				for k := range ops {
					if ops[k].Hash == origination.Hash {
						opg = ops[k]
						break
					}
				}
				if opg.Hash == "" {
					continue
				}

				if len(opg.Contents) < int(origination.ContentIndex) {
					continue
				}

				var script noderpc.Script
				if err := json.Unmarshal(opg.Contents[origination.ContentIndex].Script, &script); err != nil {
					return nil, err
				}

				storageType, err := script.Code.Storage.ToTypedAST()
				if err != nil {
					return nil, err
				}

				var storageData ast.UntypedAST
				if err := json.Unmarshal(script.Storage, &storageData); err != nil {
					return nil, err
				}

				if err := storageType.Settle(storageData); err != nil {
					return nil, err
				}

				balances, err := events.Execute(rpc, event, events.Context{
					Network:                  tzip.Network,
					Parameters:               storageType,
					Source:                   origination.Source,
					Initiator:                origination.Initiator,
					Amount:                   origination.Amount,
					HardGasLimitPerOperation: protocol.Constants.HardGasLimitPerOperation,
					ChainID:                  state.ChainID,
				})
				if err != nil {
					return nil, err
				}

				res, err := transferParsers.NewDefaultBalanceParser(h.TokenBalances).Parse(balances, origination)
				if err != nil {
					return nil, err
				}

				for i := range res {
					newModels = append(newModels, res[i])
				}

				for i := range balances {
					newModels = append(newModels, &tokenbalance.TokenBalance{
						Network:  tzip.Network,
						Address:  balances[i].Address,
						TokenID:  balances[i].TokenID,
						Contract: tzip.Address,
						Balance:  balances[i].Value,
					})
				}
			}
		}
	}

	return newModels, nil
}
