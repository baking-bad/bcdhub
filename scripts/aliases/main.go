package main

import (
	"fmt"
	"time"

	"github.com/schollz/progressbar/v3"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/tzkt"
)

const (
	dbURI = "host=127.0.0.1 port=5432 user=root dbname=bcd password=root sslmode=disable"
)

func main() {
	start := time.Now()

	api := tzkt.NewTzKT(tzkt.TzKTURLV1, time.Second*time.Duration(10))
	logger.Success("Initialized tzkt api [%s]", tzkt.TzKTURLV1)

	db, err := database.New(dbURI)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	logger.Success("Initialized database [%s]", dbURI)

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
