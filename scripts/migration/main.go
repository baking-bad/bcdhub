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
	&migrations.CreateTransfersTags{},
	&migrations.CreateTZIP{},
	&migrations.FillTZIP{},
	&migrations.ExtendedStorageEvents{},
	&migrations.ParameterEvents{},
	&migrations.TokenBalanceRecalc{},
	&migrations.NFTMetadata{},
	&migrations.TokenMetadataUnknown{},
	&migrations.DefaultEntrypoint{},
	&migrations.FixLostSearchContracts{},
	&migrations.FixEntrypointSearch{},
	&migrations.FindLostContracts{},
	&migrations.BigMapStateCount{},
}

func main() {
	migration, err := chooseMigration()
	if err != nil {
		logger.Err(err)
		return
	}

	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Err(err)
		return
	}

	start := time.Now()

	ctxs := config.NewContexts(
		cfg, cfg.Scripts.Networks,
		config.WithStorage(cfg.Storage, "migrations", 0, cfg.Scripts.Connections.Open, cfg.Scripts.Connections.Idle, false),
		config.WithRPC(cfg.RPC, false),
		config.WithConfigCopy(cfg),
		config.WithLoadErrorDescriptions(),
		config.WithSearch(cfg.Storage),
	)
	defer ctxs.Close()

	for _, ctx := range ctxs {
		logger.Info().Msgf("Starting %v migration for %s...", migration.Key(), ctx.Network.String())
		if err := migration.Do(ctx); err != nil {
			logger.Err(err)
			return
		}
	}

	logger.Info().Msgf("%s migration done. Spent: %v", migration.Key(), time.Since(start))
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
