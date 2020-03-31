package migrations

import (
	"log"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/metrics"
)

// SetOperationAliasMigration - migration that set source or destination alias from db to operations in mainnet
type SetOperationAliasMigration struct{}

// Do - migrate function
func (m *SetOperationAliasMigration) Do(ctx *Context) error {
	start := time.Now()
	h := metrics.New(ctx.ES, ctx.DB)

	operations, err := ctx.ES.GetAllOperations(consts.Mainnet)
	if err != nil {
		return err
	}

	for i := range operations {
		if err := h.SetOperationAliases(&operations[i]); err != nil {
			return err
		}

		if _, err := ctx.ES.UpdateDoc(elastic.DocOperations, operations[i].ID, operations[i]); err != nil {
			return err
		}

		log.Printf("Done %d/%d", i, len(operations))
	}

	log.Printf("Time spent: %v", time.Since(start))

	return nil
}
