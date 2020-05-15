package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/scripts/migration/migrations"
)

func main() {
	migrationMap := map[string]migrations.Migration{
		"timestamp":                 &migrations.SetTimestamp{},
		"language":                  &migrations.SetLanguage{},
		"contract_alias":            &migrations.SetContractAlias{Network: consts.Mainnet},
		"operation_alias":           &migrations.SetOperationAlias{Network: consts.Mainnet},
		"bmd_strings":               &migrations.SetBMDStrings{},
		"bmd_timestamp":             &migrations.SetBMDTimestamp{},
		"fa1_tag":                   &migrations.SetFA1{},
		"operation_strings":         &migrations.SetOperationStrings{},
		"operation_burned":          &migrations.SetOperationBurned{},
		"total_withdrawn":           &migrations.SetTotalWithdrawn{},
		"set_migration_kind":        &migrations.SetMigrationKind{},
		"set_bmd_protocol":          &migrations.SetBMDProtocol{},
		"lost":                      &migrations.FindLostOperations{},
		"contract_migrations_count": &migrations.SetContractMigrationsCount{},
		"state_chain_id":            &migrations.SetStateChainID{},
		"set_alias_slug":            &migrations.SetAliasSlug{},
		"set_contract_fingerprint":  &migrations.SetContractFingerprint{},
		"set_operation_errors":      &migrations.SetOperationErrors{},
		"set_contract_hash":         &migrations.SetContractHash{},
		"recalc_contract_metrics":   &migrations.RecalcContractMetrics{},
	}

	migrationName, err := chooseMigration(migrationMap)
	if err != nil {
		logger.Fatal(err)
	}

	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	start := time.Now()

	ctx := config.NewContext(
		config.WithElasticSearch(cfg.Elastic),
		config.WithDatabase(cfg.DB),
		config.WithRPC(cfg.RPC),
		config.WithConfigCopy(cfg),
		config.WithLoadErrorDescriptions("data/errors.json"),
	)
	defer ctx.Close()

	logger.Info("Starting %v migration...", migrationName)

	if err := migrationMap[migrationName].Do(ctx); err != nil {
		log.Fatal(err)
	}

	logger.Success("%s migration done. Spent: %v", migrationName, time.Since(start))
}

func chooseMigration(migrationMap map[string]migrations.Migration) (string, error) {
	var mKeys = make([]string, 0, len(migrationMap))

	for m := range migrationMap {
		mKeys = append(mKeys, m)
	}

	sort.Strings(mKeys)

	fmt.Println("Available migrations:")
	for i, name := range mKeys {
		spaces := 30 - len(name)
		desc := migrationMap[name].Description()

		if i > 9 {
			spaces--
		}

		fmt.Printf("[%d] %s%s| %s\n", i, name, strings.Repeat(" ", spaces), desc)
	}

	var input int
	fmt.Println("\nEnter migration #:")
	fmt.Scanln(&input)

	if input < 0 || input > len(mKeys)-1 {
		return "", fmt.Errorf("Invalid # of migration: %d", input)
	}

	migration := mKeys[input]
	if _, ok := migrationMap[migration]; !ok {
		return "", fmt.Errorf("Unknown migration key: %s", migration)
	}

	return migration, nil
}
