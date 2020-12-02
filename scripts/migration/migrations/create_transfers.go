package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/parsers/stacktrace"
	"github.com/baking-bad/bcdhub/internal/parsers/transfer"
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

	h := metrics.New(ctx.ES, ctx.DB)

	operations, err := m.getOperations(ctx)
	if err != nil {
		return err
	}
	logger.Info("Found %d operations with transfer entrypoint", len(operations))

	result := make([]elastic.Model, 0)

	bar := progressbar.NewOptions(len(operations), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
	for i := range operations {
		if err := bar.Add(1); err != nil {
			return err
		}
		rpc, err := ctx.GetRPC(operations[i].Network)
		if err != nil {
			return err
		}

		protocol, err := ctx.ES.GetProtocol(operations[i].Network, "", -1)
		if err != nil {
			return err
		}

		parser, err := transfer.NewParser(rpc, ctx.ES,
			transfer.WithNetwork(operations[i].Network),
			transfer.WithGasLimit(protocol.Constants.HardGasLimitPerOperation),
			transfer.WithStackTrace(stacktrace.New()),
			transfer.WithoutViews(),
		)
		if err != nil {
			return err
		}

		transfers, err := parser.Parse(operations[i], nil)
		if err != nil {
			return err
		}

		for j := range transfers {
			h.SetTransferAliases(ctx.Aliases, transfers[j])
			// logger.Info("%s %##v", operations[i].Entrypoint, transfers[j])
			result = append(result, transfers[j])
		}
	}

	if err := ctx.ES.BulkInsert(result); err != nil {
		logger.Errorf("ctx.ES.BulkUpdate error: %v", err)
		return err
	}

	logger.Info("Done. %d transfers were saved", len(result))

	return nil
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

	return ctx.ES.DeleteByContract([]string{elastic.DocTransfers}, m.Network, m.Address)
}

func (m *CreateTransfersTags) getOperations(ctx *config.Context) ([]models.Operation, error) {
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
	return ctx.ES.GetOperations(filters, 0, false)
}
