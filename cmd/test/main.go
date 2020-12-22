package main

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
)

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	ctx := config.NewContext(
		config.WithElasticSearch(cfg.Elastic),
	)
	defer ctx.Close()

	network := consts.Mainnet
	contracts := []string{"KT1LN4LPSqTMS7Sd2CJw4bbDGRkMv2t68Fy9"}

	start := time.Now().Add(time.Duration(-24)*time.Hour).Unix() * 1000
	end := time.Now().Unix() * 1000
	tokenID := int64(0)

	transfers, err := ctx.ES.GetTransfers(elastic.GetTransfersContext{
		Network:   network,
		Contracts: contracts,
		Start:     uint(start),
		End:       uint(end),
		TokenID:   tokenID,
	})

	var total float64
	for _, t := range transfers.Transfers {
		total += t.Amount
	}

	fmt.Println("total", transfers.Total, "amount", total)
}
