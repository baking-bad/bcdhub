package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/scripts/migration/migrations"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

var migrationsList = []migrations.Migration{
	&migrations.SetTimestamp{},
	&migrations.SetLanguage{},
	&migrations.SetContractAlias{Network: consts.Mainnet},
	&migrations.SetOperationAlias{Network: consts.Mainnet},
	&migrations.SetBMDStrings{},
	&migrations.SetBMDTimestamp{},
	&migrations.SetFA{},
	&migrations.SetOperationStrings{},
	&migrations.SetTotalWithdrawn{},
	&migrations.SetMigrationKind{},
	&migrations.SetBMDProtocol{},
	&migrations.FindLostOperations{},
	&migrations.SetContractMigrationsCount{},
	&migrations.SetStateChainID{},
	&migrations.SetAliasSlug{},
	&migrations.SetContractFingerprint{},
	&migrations.SetOperationErrors{},
	&migrations.SetContractHash{},
	&migrations.RecalcContractMetrics{},
	&migrations.SetEmptyBmdPtr{},
	&migrations.DropMichelson{},
	&migrations.SetOperationTags{},
	&migrations.CreateTransfersTags{},
	&migrations.SetProtocolConstants{},
	&migrations.SetOperationAllocatedBurned{},
	&migrations.CreateTokenMetadata{},
	&migrations.SetOperationInitiator{},
	&migrations.UpdateDapps{},
	&migrations.CreateTZIP{},
}

func main() {
	migration, err := chooseMigration()
	if err != nil {
		logger.Fatal(err)
	}

	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	gjson.AddModifier("upper", func(json, arg string) string {
		return strings.ToUpper(json)
	})
	gjson.AddModifier("lower", func(json, arg string) string {
		return strings.ToLower(json)
	})

	start := time.Now()

	ctx := config.NewContext(
		config.WithElasticSearch(cfg.Elastic),
		config.WithDatabase(cfg.DB),
		config.WithRPC(cfg.RPC),
		config.WithConfigCopy(cfg),
		config.WithLoadErrorDescriptions("data/errors.json"),
		config.WithContractsInterfaces(),
		config.WithAliases(consts.Mainnet),
	)
	defer ctx.Close()

	logger.Info("Starting %v migration...", migration.Key())

	if err := migration.Do(ctx); err != nil {
		log.Fatal(err)
	}

	logger.Success("%s migration done. Spent: %v", migration.Key(), time.Since(start))
}

func chooseMigration() (migrations.Migration, error) {
	fmt.Println("Available migrations:")
	for i, migration := range migrationsList {
		spaces := 30 - len(migration.Key()) - int(math.Log10(float64(i)+0.1))
		desc := migration.Description()

		fmt.Printf("[%d] %s%s| %s\n", i, migration.Key(), strings.Repeat(" ", spaces), desc)
	}

	var input string
	fmt.Println("\nEnter migration #:")
	fmt.Scanln(&input)

	index, err := strconv.Atoi(input)
	if err != nil {
		return nil, err
	}

	if index < 0 || index > len(migrationsList)-1 {
		return nil, errors.Errorf("Invalid # of migration: %s", input)
	}

	return migrationsList[index], nil
}
