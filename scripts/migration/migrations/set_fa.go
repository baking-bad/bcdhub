package migrations

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/schollz/progressbar/v3"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/kinds"
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

	for name, entrypoints := range ctx.Interfaces {
		logger.Info("Found %s interface", name)

		updates := make([]elastic.Model, 0)
		bar := progressbar.NewOptions(len(contracts), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish())
		for _, c := range contracts {
			bar.Add(1)

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
			if !findInterface(m, entrypoints) {
				continue
			}

			c.Tags = append(c.Tags, name)
			updates = append(updates, &c)
		}

		if err := ctx.ES.BulkUpdate(updates); err != nil {
			logger.Errorf("ctx.ES.BulkUpdate error: %v", err)
			return err
		}
	}

	return nil
}

func findInterface(metadata meta.Metadata, i []kinds.Entrypoint) bool {
	root := metadata["0"]

	for _, ie := range i {
		found := false
		for _, e := range root.Args {
			entrypointMeta := metadata[e]
			if compareEntrypoints(metadata, ie, *entrypointMeta, e) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func compareEntrypoints(metadata meta.Metadata, in kinds.Entrypoint, en meta.NodeMetadata, path string) bool {
	if in.Name != "" && en.Name != in.Name {
		return false
	}
	// fmt.Printf("[in] %+v\n[en] %+v\n\n", in, en)
	if in.Prim != en.Prim {
		return false
	}

	for i, inArg := range in.Args {
		enPath := fmt.Sprintf("%s/%d", path, i)
		enMeta, ok := metadata[enPath]
		if !ok {
			return false
		}
		if !compareEntrypoints(metadata, inArg, *enMeta, enPath) {
			return false
		}
	}

	return true
}
