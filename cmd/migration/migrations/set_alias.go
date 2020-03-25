package migrations

import (
	"log"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/elastic"
)

// SetAliasMigration - migration that set source or destination alias from db to operations in mainnet
type SetAliasMigration struct{}

// Do - migrate function
func (m *SetAliasMigration) Do(ctx *Context) error {
	start := time.Now()

	operations, err := ctx.ES.GetAllOperations(consts.Mainnet)
	if err != nil {
		return err
	}

	aliasesFromDB, err := ctx.DB.GetAliases(consts.Mainnet)
	if err != nil {
		return err
	}

	aliases := make(map[string]string, len(aliasesFromDB))
	for _, a := range aliasesFromDB {
		aliases[a.Address] = a.Alias
	}

	var countSrc int
	var countDst int

	for i, operation := range operations {
		var found bool

		if operation.SourceAlias == "" {
			if alias, ok := aliases[operation.Source]; ok {
				operation.SourceAlias = alias
				found = true
				countSrc++
				// log.Printf("src [%v] %v", operation.Source, alias)
			}
		}

		if operation.DestinationAlias == "" {
			if alias, ok := aliases[operation.Destination]; ok {
				operation.DestinationAlias = alias
				found = true
				countDst++
				// log.Printf("dst [%v] %v", operation.Destination, alias)
			}
		}

		if !found {
			continue
		}

		if _, err := ctx.ES.UpdateDoc(elastic.DocOperations, operation.ID, operation); err != nil {
			return err
		}

		log.Printf("Done %d/%d", i, len(operations))
	}

	log.Printf("Total operations in elastic [%v]: %v", consts.Mainnet, len(operations))
	log.Printf("Total aliases in postgres [%v]: %v", consts.Mainnet, len(aliasesFromDB))
	log.Printf("Updated %v source aliases and %v destination aliases", countSrc, countDst)
	log.Printf("Time spent: %v", time.Since(start))

	return nil
}
