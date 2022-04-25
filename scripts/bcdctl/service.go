package main

import (
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
)

type listServicesCommand struct{}

var listServicesCmd listServicesCommand

// Execute
func (x *listServicesCommand) Execute(_ []string) error {
	for network, ctx := range ctxs {
		logger.Info().Msgf("----------------%s----------------", network.String())

		states, err := ctx.Services.All()
		if err != nil {
			return err
		}

		ids := make(map[string]int64)

		for _, s := range states {
			fields := map[string]interface{}{
				"current": s.LastID,
			}

			var lastID int64
			query := ctx.StorageDB.DB.Model()
			switch s.Name {
			case "projects", "contracts":
				if id, ok := ids[models.DocContracts]; ok {
					lastID = id
				} else {
					query.Table(models.DocContracts)
				}
			case "contract_metadata", "token_metadata", "big_map_diffs":
				if id, ok := ids[models.DocBigMapDiff]; ok {
					lastID = id
				} else {
					query.Table(models.DocBigMapDiff)
				}
			case "operations":
				if id, ok := ids[models.DocOperations]; ok {
					lastID = id
				} else {
					query.Table(models.DocOperations)
				}
			}

			if lastID == 0 {
				if err := query.ColumnExpr("max(id)").Select(&lastID); err != nil {
					return err
				}
			}

			fields["last"] = lastID
			fields["to-do"] = lastID - s.LastID
			logger.Info().Fields(fields).Msg(s.Name)
		}
	}
	return nil
}
