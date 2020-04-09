package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/jsonload"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/scripts/migration/migrations"
)

func main() {
	migrationMap := map[string]migrations.Migration{
		"timestamp":       &migrations.SetTimestampMigration{},
		"language":        &migrations.SetLanguageMigration{},
		"contract_alias":  &migrations.SetContractAliasMigration{Network: consts.Mainnet},
		"operation_alias": &migrations.SetOperationAliasMigration{Network: consts.Mainnet},
		"bmd_key_strings": &migrations.SetBMDKeyStrings{},
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

	start := time.Now()

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

	logger.Info("Starting %v migration...", env)

	if err := migrationMap[env].Do(ctx); err != nil {
		log.Fatal(err)
	}

	logger.Success("%s migration done. Spent: %v", env, time.Since(start))
}
