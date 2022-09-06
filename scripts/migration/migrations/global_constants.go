package migrations

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	astContract "github.com/baking-bad/bcdhub/internal/bcd/contract"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/contract"
)

// GlobalConstantsRelations -
type GlobalConstantsRelations struct{}

// Key -
func (m *GlobalConstantsRelations) Key() string {
	return "recover_global_constants"
}

// Description -
func (m *GlobalConstantsRelations) Description() string {
	return "recover `global_constants` relation"
}

// Do - migrate function
func (m *GlobalConstantsRelations) Do(ctx *config.Context) error {
	var offset int
	var end bool
	for !end {
		scripts, err := ctx.Scripts.GetScripts(10, offset)
		if err != nil {
			if strings.Contains(err.Error(), "no rows in result set") {
				end = true
				continue
			}
			return err
		}

		for i := range scripts {
			var accountID int64
			if err := ctx.StorageDB.DB.Model(&contract.Contract{}).
				Column("account_id").
				Where("jakarta_id = ?", scripts[i].ID).
				WhereOr("babylon_id = ?", scripts[i].ID).
				OrderExpr("id ASC").Limit(1).
				Select(&accountID); err != nil {
				if strings.Contains(err.Error(), "no rows in result set") {
					continue
				}
				return err
			}

			var address string
			if err := ctx.StorageDB.DB.Model(&account.Account{}).
				Column("address").
				Where("id = ?", accountID).
				OrderExpr("id ASC").Limit(1).
				Select(&address); err != nil {
				if strings.Contains(err.Error(), "no rows in result set") {
					continue
				}
				return err
			}

			logger.Info().Str("address", address).Msg("finding constants...")

			data, err := ctx.RPC.GetRawScript(context.Background(), address, 0)
			if err != nil {
				return err
			}

			var cd astContract.ContractData
			if err := json.Unmarshal(data, &cd); err != nil {
				return err
			}

			var tree ast.UntypedAST
			if err := json.Unmarshal(cd.Code, &tree); err != nil {
				return err
			}

			constants, err := astContract.FindConstants(tree)
			if err != nil {
				return err
			}

			if len(constants) == 0 {
				continue
			}
			logger.Info().Str("address", address).Int("constants_count", len(constants)).Msg("found constants")

			for key := range constants {
				gc, err := ctx.GlobalConstants.Get(key)
				if err != nil {
					return err
				}
				relation := contract.ScriptConstants{
					GlobalConstantId: gc.ID,
					ScriptId:         scripts[i].ID,
				}

				if _, err := ctx.StorageDB.DB.Model(&relation).Insert(); err != nil {
					return err
				}
			}
		}

		offset += len(scripts)
		end = len(scripts) == 0
	}
	return nil
}
