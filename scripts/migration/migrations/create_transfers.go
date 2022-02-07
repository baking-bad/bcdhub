package migrations

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	transferParsers "github.com/baking-bad/bcdhub/internal/parsers/transfer"
	"github.com/schollz/progressbar/v3"
)

// CreateTransfersTags -
type CreateTransfersTags struct {
	Network string
	Address string
}

// Key -
func (m *CreateTransfersTags) Key() string {
	return "create_transfers"
}

// Description -
func (m *CreateTransfersTags) Description() string {
	return "creates 'transfer' index"
}

// Do - migrate function
func (m *CreateTransfersTags) Do(ctx *config.Context) error {
	logger.Info().Msg("Starting create transfer migration...")
	if err := m.deleteTransfers(ctx); err != nil {
		return err
	}

	operations, err := m.getOperations(ctx)
	if err != nil {
		return err
	}
	logger.Info().Msgf("Found %d operations with transfer entrypoint", len(operations))

	result := make([]models.Model, 0)
	newTransfers := make([]*transfer.Transfer, 0)
	bar := progressbar.NewOptions(len(operations), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
	for i := range operations {
		if err := bar.Add(1); err != nil {
			return err
		}
		rpc, err := ctx.GetRPC(operations[i].Network)
		if err != nil {
			return err
		}

		protocol, err := ctx.Protocols.Get(operations[i].Network, "", -1)
		if err != nil {
			return err
		}

		parser, err := transferParsers.NewParser(rpc, ctx.ContractMetadata, ctx.Blocks, ctx.TokenBalances, ctx.Accounts,
			transferParsers.WithNetwork(operations[i].Network),
			transferParsers.WithGasLimit(protocol.Constants.HardGasLimitPerOperation),
			transferParsers.WithoutViews(),
		)
		if err != nil {
			return err
		}
		proto, err := ctx.Cache.ProtocolByID(operations[i].Network, operations[i].ProtocolID)
		if err != nil {
			return err
		}
		script, err := ctx.Contracts.Script(operations[i].Network, operations[i].Destination.Address, proto.SymLink)
		if err != nil {
			return err
		}
		operations[i].Script, err = script.Full()
		if err != nil {
			return err
		}
		operations[i].Script = script.Code

		if err := parser.Parse(nil, proto.Hash, &operations[i]); err != nil {
			return err
		}

		for j := range operations[i].Transfers {
			result = append(result, operations[i].Transfers[j])
			newTransfers = append(newTransfers, operations[i].Transfers[j])
		}
	}

	balanceUpdates := transferParsers.UpdateTokenBalances(newTransfers)
	for i := range balanceUpdates {
		result = append(result, balanceUpdates[i])
	}

	return ctx.Storage.Save(context.Background(), result)
}

func (m *CreateTransfersTags) deleteTransfers(ctx *config.Context) (err error) {
	m.Network, err = ask("Enter network (empty if all):")
	if err != nil {
		return
	}

	typ := types.Empty
	if m.Network != "" {
		if m.Address, err = ask("Enter KT address (empty if all):"); err != nil {
			return
		}
		typ = types.NewNetwork(m.Network)
	}

	return ctx.Storage.DeleteByContract(typ, []string{models.DocTransfers}, m.Address)
}

func (m *CreateTransfersTags) getOperations(ctx *config.Context) ([]operation.Operation, error) {
	filters := map[string]interface{}{}
	if m.Network != "" {
		filters["operation.network"] = m.Network
		if m.Address != "" {
			filters["destination.address"] = m.Address
		} else {
			filters["entrypoint"] = "transfer"
		}
	} else {
		filters["entrypoint"] = "transfer"
	}
	return ctx.Operations.Get(filters, 0, false)
}
