package metrics

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/events"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	transferParsers "github.com/baking-bad/bcdhub/internal/parsers/transfer"
	"github.com/baking-bad/bcdhub/internal/parsers/tzip/tokens"
	"github.com/pkg/errors"
)

// CreateTokenMetadata -
func (h *Handler) CreateTokenMetadata(rpc noderpc.INode, sharePath string, c *contract.Contract, ipfs ...string) error {

	result := make([]models.Model, 0)

	transfers, err := h.ExecuteInitialStorageEvent(rpc, c.Network, c.Address)
	if err != nil {
		return err
	}
	for i := range transfers {
		result = append(result, transfers[i])
	}

	return h.Storage.BulkInsert(result)
}

// FixTokenMetadata -
func (h *Handler) FixTokenMetadata(rpc noderpc.INode, sharePath string, contract *contract.Contract, operation *operation.Operation, ipfs ...string) error {
	if !operation.IsTransaction() || !operation.IsApplied() || !operation.IsCall() {
		return nil
	}

	tokenMetadatas, err := h.TokenMetadata.Get(tokenmetadata.GetContext{
		Contract: operation.Destination,
		Network:  operation.Network,
		TokenID:  -1,
	})
	if err != nil {
		if !h.Storage.IsRecordNotFound(err) {
			return err
		}
		return nil
	}
	result := make([]models.Model, 0)

	for _, tokenMetadata := range tokenMetadatas {
		parser := tokens.NewParser(h.BigMapDiffs, h.Blocks, h.Protocol, h.Storage, rpc, sharePath, operation.Network, ipfs...)
		metadata, err := parser.Parse(tokenMetadata.Contract, operation.Level)
		if err != nil {
			return err
		}
		for i := range metadata {
			result = append(result, &metadata[i])
			logger.With(&metadata[i]).Info("Token metadata update is found")
		}
	}
	if len(result) == 0 {
		return nil
	}

	return h.Storage.BulkInsert(result)
}

// ExecuteInitialStorageEvent -
func (h *Handler) ExecuteInitialStorageEvent(rpc noderpc.INode, network, contract string) ([]*transfer.Transfer, error) {
	tzip, err := h.TZIP.Get(network, contract)
	if err != nil {
		if h.Storage.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	ops, err := h.Operations.Get(map[string]interface{}{
		"destination": contract,
		"network":     network,
		"kind":        consts.Origination,
	}, 1, false)
	if err != nil {
		if h.Storage.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(ops) != 1 {
		return nil, errors.Errorf("Invalid operations count: len(ops) [%d] != 1", len(ops))
	}

	origination := ops[0]

	protocol, err := h.Protocol.GetProtocol(network, origination.Protocol, origination.Level)
	if err != nil {
		return nil, err
	}

	state, err := h.Blocks.Last(network)
	if err != nil {
		return nil, err
	}

	transfers := make([]*transfer.Transfer, 0)

	balanceUpdates := make([]*tokenbalance.TokenBalance, 0)
	for i := range tzip.Events {
		for j := range tzip.Events[i].Implementations {
			if !tzip.Events[i].Implementations[j].MichelsonInitialStorageEvent.Empty() {
				event, err := events.NewMichelsonInitialStorage(tzip.Events[i].Implementations[j], tzip.Events[i].Name)
				if err != nil {
					return nil, err
				}

				ops, err := rpc.GetOperations(origination.Level)
				if err != nil {
					return nil, err
				}

				path := fmt.Sprintf(`#(hash=="%s")#.contents.%d.script.storage`, origination.Hash, origination.ContentIndex)
				defattedStorage := ops.Get(path).Array()
				if len(defattedStorage) == 0 {
					return nil, fmt.Errorf("[ExecuteInitialStorageEvent] Empty storage")
				}

				balances, err := events.Execute(rpc, event, events.Context{
					Network:                  tzip.Network,
					Parameters:               defattedStorage[0].String(),
					Source:                   origination.Source,
					Initiator:                origination.Initiator,
					Amount:                   origination.Amount,
					HardGasLimitPerOperation: protocol.Constants.HardGasLimitPerOperation,
					ChainID:                  state.ChainID,
				})
				if err != nil {
					return nil, err
				}

				res, err := transferParsers.NewDefaultBalanceParser(h.Storage).Parse(balances, origination)
				if err != nil {
					return nil, err
				}

				transfers = append(transfers, res...)

				for i := range balances {
					balanceUpdates = append(balanceUpdates, &tokenbalance.TokenBalance{
						Network:  tzip.Network,
						Address:  balances[i].Address,
						TokenID:  balances[i].TokenID,
						Contract: tzip.Address,
						Value:    balances[i].Value,
					})
				}
			}
		}
	}

	return transfers, h.TokenBalances.Update(balanceUpdates)
}
