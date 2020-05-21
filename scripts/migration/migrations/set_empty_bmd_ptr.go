package migrations

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/contractparser/newmiguel"
	"github.com/baking-bad/bcdhub/internal/contractparser/storage"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/schollz/progressbar/v3"
)

// SetEmptyBmdPtr - migration that set empty big map diff ptr (for alpha protocol)
type SetEmptyBmdPtr struct {
	Network string
}

// Key -
func (m *SetEmptyBmdPtr) Key() string {
	return "set_empty_bmd_ptr"
}

// Description -
func (m *SetEmptyBmdPtr) Description() string {
	return "set empty big map diff ptr (for alpha protocol)"
}

// Do - migrate function
func (m *SetEmptyBmdPtr) Do(ctx *config.Context) error {
	empty, err := ctx.ES.GetBigMapDiffsWithEmptyPtr()
	if err != nil {
		return err
	}
	logger.Info("Found %d big map diffs with empty pointers", len(empty))

	pointerMap := make(map[string]int64)

	updates := make([]elastic.Model, len(empty))
	bar := progressbar.NewOptions(len(empty), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish())
	for i := range empty {
		bar.Add(1)

		rpc, err := ctx.GetRPC(empty[i].Network)
		if err != nil {
			return err
		}

		if ptr, ok := pointerMap[empty[i].Address]; ok && empty[i].BinPath == "0/0" && empty[i].Network == "mainnet" {
			empty[i].Ptr = ptr
		} else {
			storageJSON, err := rpc.GetScriptStorageJSON(empty[i].Address, 0)
			if err != nil {
				return err
			}

			metadata, err := meta.GetMetadata(ctx.ES, empty[i].Address, "storage", "PsBabyM1eUXZseaJdmXFApDSBqj8YBfwELoxZHHW77EMcAbbwAS")
			if err != nil {
				return err
			}

			binPathMap, err := storage.FindBigMapPointers(metadata, storageJSON)
			if err != nil {
				return err
			}

			if len(binPathMap) != 1 {
				return fmt.Errorf("Invalid big map diff counter: %d", len(binPathMap))
			}

			for pointer, binPath := range binPathMap {
				path := newmiguel.GetGJSONPath(strings.TrimPrefix(binPath, "0/"))
				path += ".int"
				empty[i].BinPath = binPath
				empty[i].Ptr = pointer
				pointerMap[empty[i].Address] = pointer
			}
		}
		updates[i] = &empty[i]
	}

	if err := ctx.ES.BulkUpdate(updates); err != nil {
		return err
	}

	return nil
}
