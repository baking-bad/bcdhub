package migrations

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/schollz/progressbar/v3"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/elastic"
)

// SetFA - migration that set fa tag to contract
type SetFA struct{}

// Key -
func (m *SetFA) Key() string {
	return "fa_tag"
}

// Description -
func (m *SetFA) Description() string {
	return "set fa tag to contract"
}

// Do - migrate function
func (m *SetFA) Do(ctx *config.Context) error {
	contracts, err := ctx.ES.GetContracts(nil)
	if err != nil {
		return err
	}

	logger.Info("Found %d contracts", len(contracts))
	updates := make([]elastic.Model, 0)

	bar := progressbar.NewOptions(len(contracts), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish())
	for _, c := range contracts {
		bar.Add(1) //nolint

		m, err := meta.GetMetadata(ctx.ES, c.Address, consts.PARAMETER, "PsBabyM1eUXZseaJdmXFApDSBqj8YBfwELoxZHHW77EMcAbbwAS")
		if err != nil {
			if !strings.Contains(err.Error(), "Unknown metadata sym link") {
				return err
			}
			m, err = meta.GetMetadata(ctx.ES, c.Address, consts.PARAMETER, "PtYuensgYBb3G3x1hLLbCmcav8ue8Kyd2khADcL5LsT5R1hcXex")
			if err != nil {
				return err
			}
		}

		parameter := new(contractparser.Parameter)
		parameter.Tags = make(helpers.Set)
		parameter.Metadata = m

		if err := parameter.FindTags(ctx.Interfaces); err != nil {
			return err
		}

		c.Tags = append(c.Tags, parameter.Tags.Values()...)
		updates = append(updates, &c)
	}

	if err := ctx.ES.BulkUpdate(updates); err != nil {
		logger.Errorf("ctx.ES.BulkUpdate error: %v", err)
		return err
	}

	return nil
}
