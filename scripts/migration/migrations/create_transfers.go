package migrations

import (
	"bufio"
	"os"
	"strings"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/parsers"
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

	tokenViews, err := parsers.NewTokenViews(ctx.DB)
	if err != nil {
		return err
	}

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

		parser := parsers.NewTransferParser(rpc, ctx.ES, parsers.WithTokenViewsTransferParser(tokenViews))

		transfers, err := parser.Parse(operations[i])
		if err != nil {
			return err
		}

		for i := range transfers {
			if _, err := h.SetTransferAliases(ctx.Aliases, transfers[i]); err != nil {
				return err
			}
			result = append(result, transfers[i])

		}
	}

	if err := ctx.ES.BulkInsert(result); err != nil {
		logger.Errorf("ctx.ES.BulkUpdate error: %v", err)
		return err
	}

	logger.Info("Done. %d transfers were saved", len(result))

	return nil
}

func (m *CreateTransfersTags) ask(question string) (string, error) {
	logger.Question(question)
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.Replace(text, "\n", "", -1), nil
}

func (m *CreateTransfersTags) deleteTransfers(ctx *config.Context) (err error) {
	m.Network, err = m.ask("Enter network (empty if all):")
	if err != nil {
		return
	}
	if m.Network != "" {
		if m.Address, err = m.ask("Enter KT address (empty if all):"); err != nil {
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
