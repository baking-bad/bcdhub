package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/scripts/migration/migrations"
	"github.com/pkg/errors"
)

var migrationsList = []migrations.Migration{
	&migrations.BigRussianBoss{},
	&migrations.GetAliases{},
	&migrations.SetAliases{},
	&migrations.RecalcContractMetrics{},
	&migrations.SetOperationTags{},
	&migrations.CreateTransfersTags{},
	&migrations.SetProtocolConstants{},
	&migrations.CreateTZIP{},
	&migrations.FillTZIP{},
	&migrations.InitialStorageEvents{},
	&migrations.ExtendedStorageEvents{},
	&migrations.ParameterEvents{},
	&migrations.TokenBalanceRecalc{},
	&migrations.TokenMetadataSetDecimals{},
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

	start := time.Now()

	ctx := config.NewContext(
		config.WithStorage(cfg.Storage),
		config.WithDatabase(cfg.DB),
		config.WithRPC(cfg.RPC),
		config.WithConfigCopy(cfg),
		config.WithLoadErrorDescriptions(),
	)
	defer ctx.Close()

	logger.Info("Starting %v migration...", migration.Key())

	if err := migration.Do(ctx); err != nil {
		logger.Fatal(err)
	}

	logger.Info("%s migration done. Spent: %v", migration.Key(), time.Since(start))
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
