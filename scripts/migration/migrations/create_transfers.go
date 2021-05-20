package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/fetch"
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
	logger.Info("Starting create transfer migration...")
	if err := m.deleteTransfers(ctx); err != nil {
		return err
	}

	operations, err := m.getOperations(ctx)
	if err != nil {
		return err
	}
	logger.Info("Found %d operations with transfer entrypoint", len(operations))

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

		parser, err := transferParsers.NewParser(rpc, ctx.TZIP, ctx.Blocks, ctx.Storage,
			ctx.SharePath,
			transferParsers.WithNetwork(operations[i].Network),
			transferParsers.WithGasLimit(protocol.Constants.HardGasLimitPerOperation),
			transferParsers.WithoutViews(),
		)
		if err != nil {
			return err
		}
		operations[i].Script, err = fetch.Contract(operations[i].Network, operations[i].Destination, operations[i].Protocol, ctx.SharePath)
		if err != nil {
			return err
		}
		transfers, err := parser.Parse(operations[i], nil)
		if err != nil {
			return err
		}

		for j := range transfers {
			result = append(result, transfers[j])
			newTransfers = append(newTransfers, transfers[j])
		}
	}

	balanceUpdates := transferParsers.UpdateTokenBalances(newTransfers)
	for i := range balanceUpdates {
		result = append(result, balanceUpdates[i])
	}

	return ctx.Storage.Save(result)
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
		filters["network"] = m.Network
		if m.Address != "" {
			filters["destination"] = m.Address
		} else {
			filters["entrypoint"] = "transfer"
		}
	} else {
		filters["entrypoint"] = "transfer"
	}
	return ctx.Operations.Get(filters, 0, false)
}
