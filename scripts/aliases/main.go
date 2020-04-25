package main

import (
	"fmt"
	"time"

	"github.com/schollz/progressbar/v3"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/tzkt"
)

func main() {
	start := time.Now()
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	timeout := time.Duration(cfg.TzKT[consts.Mainnet].Timeout) * time.Second
	api := tzkt.NewTzKT(cfg.TzKT[consts.Mainnet].URI, timeout)
	logger.Success("TzKT API initialized")

	db, err := database.New(cfg.DB.ConnString)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	logger.Success("Database connection established")

	aliases, err := api.GetAliases()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("Got %d aliases from tzkt api", len(aliases))

	bar := progressbar.NewOptions(len(aliases), progressbar.OptionSetPredictTime(false))

	logger.Info("Saving aliases to db...")
	for _, alias := range aliases {
		bar.Add(1)

		dbAlias := database.Alias{
			Alias:   alias.Alias,
			Network: consts.Mainnet,
			Address: alias.Address,
		}

		if err := db.CreateOrUpdateAlias(&dbAlias); err != nil {
			fmt.Print("\033[2K\r")
			logger.Fatal(fmt.Errorf("%v in <%v> with alias <%v> got error: %v", alias.Address, consts.Mainnet, alias.Alias, err))
		}
	}

	fmt.Print("\033[2K\r")
	logger.Success("Done. Spent: %v", time.Since(start))
}
