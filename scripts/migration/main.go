package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/scripts/migration/migrations"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var migrationsList = []migrations.Migration{}

func main() {
	migration, err := chooseMigration()
	if err != nil {
		log.Err(err).Msg("choose migration")
		return
	}

	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		log.Err(err).Msg("load config")
		return
	}

	start := time.Now()

	ctxs := config.NewContexts(
		cfg, cfg.Scripts.Networks,
		config.WithStorage(cfg.Storage, "migrations", 0),
		config.WithRPC(cfg.RPC),
		config.WithConfigCopy(cfg),
		config.WithLoadErrorDescriptions(),
	)
	defer ctxs.Close()

	for _, ctx := range ctxs {
		log.Info().Msgf("Starting %v migration for %s...", migration.Key(), ctx.Network.String())
		if err := migration.Do(ctx); err != nil {
			log.Err(err).Msg("migration execution")
			return
		}
	}

	log.Info().Msgf("%s migration done. Spent: %v", migration.Key(), time.Since(start))
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
	if _, err := fmt.Scanln(&input); err != nil {
		return nil, err
	}

	index, err := strconv.Atoi(input)
	if err != nil {
		return nil, err
	}

	if index < 0 || index > len(migrationsList)-1 {
		return nil, errors.Errorf("Invalid # of migration: %s", input)
	}

	return migrationsList[index], nil
}
