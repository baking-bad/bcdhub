package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
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

		protocol, err := ctx.Protocols.GetProtocol(operations[i].Network, "", -1)
		if err != nil {
			return err
		}

		parser, err := transferParsers.NewParser(rpc, ctx.TZIP, ctx.Blocks, ctx.Storage,
			transferParsers.WithNetwork(operations[i].Network),
			transferParsers.WithGasLimit(protocol.Constants.HardGasLimitPerOperation),
			transferParsers.WithoutViews(),
		)
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

	if err := ctx.Storage.BulkInsert(result); err != nil {
		logger.Errorf("ctx.Storage.BulkInsert error: %v", err)
		return err
	}

	logger.Info("Done. %d transfers were saved", len(result))

	return transferParsers.UpdateTokenBalances(ctx.TokenBalances, newTransfers)
}

func (m *CreateTransfersTags) deleteTransfers(ctx *config.Context) (err error) {
	m.Network, err = ask("Enter network (empty if all):")
	if err != nil {
		return
	}
	if m.Network != "" {
		if m.Address, err = ask("Enter KT address (empty if all):"); err != nil {
			return
		}
	}

	return ctx.Storage.DeleteByContract([]string{models.DocTransfers}, m.Network, m.Address)
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
