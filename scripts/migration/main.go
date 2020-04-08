package main

import (
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
	}

	env := os.Getenv("MIGRATION")

	if env == "" {
		log.Fatal("MIGRATION env variable is not set.")
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
