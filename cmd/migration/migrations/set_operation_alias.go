package migrations

import (
	"log"
	"time"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/metrics"
)

// SetOperationAliasMigration - migration that set source or destination alias from db to operations in choosen network
type SetOperationAliasMigration struct {
	Network string
}

// Do - migrate function
func (m *SetOperationAliasMigration) Do(ctx *Context) error {
	start := time.Now()
	h := metrics.New(ctx.ES, ctx.DB)

	operations, err := ctx.ES.GetAllOperations(m.Network)
	if err != nil {
		return err
	}

	aliases, err := h.GetAliases(m.Network)
	if err != nil {
		return err
	}

	for i := range operations {
		h.SetOperationAliases(&operations[i], aliases)

		if _, err := ctx.ES.UpdateDoc(elastic.DocOperations, operations[i].ID, operations[i]); err != nil {
			return err
		}

		log.Printf("Done %d/%d", i, len(operations))
	}

	log.Printf("Time spent: %v", time.Since(start))

	return nil
}
