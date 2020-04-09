package main

import (
	"fmt"
	"log"
	"os"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/jsonload"
	"github.com/baking-bad/bcdhub/scripts/migration/migrations"
)

func main() {
	migrationMap := map[string]migrations.Migration{
		"timestamp":       &migrations.SetTimestampMigration{},
		"language":        &migrations.SetLanguageMigration{},
		"contract_alias":  &migrations.SetContractAliasMigration{Network: consts.Mainnet},
		"operation_alias": &migrations.SetOperationAliasMigration{Network: consts.Mainnet},
		"bmd_strings":     &migrations.SetBMDStrings{},
		"bmd_timestamp":   &migrations.SetBMDTimestamp{},
		"fa1_tag":         &migrations.SetFA1Migration{},
	}

	env := os.Getenv("MIGRATION")

	if env == "" {
		fmt.Println("Set MIGRATION env variable. Available migrations:")
		for key := range migrationMap {
			fmt.Println("-", key)
		}
		return
	}

	if _, ok := migrationMap[env]; !ok {
		log.Fatal("Unknown migration key: ", env)
	}

	var cfg migrations.Config
	if err := jsonload.StructFromFile("config.json", &cfg); err != nil {
		log.Fatal(err)
	}
	cfg.Print()

	ctx, err := migrations.NewContext(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Close()

	if err := migrationMap[env].Do(ctx); err != nil {
		log.Fatal(err)
	}
}
