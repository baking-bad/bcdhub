package main

import (
	"fmt"
	"os"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
)

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	ctx := config.NewContext(
		config.WithElasticSearch(cfg.Elastic),
		config.WithConfigCopy(cfg),
	)
	defer ctx.Close()

	dapps, err := ctx.ES.GetDApps()
	if err != nil {
		logger.Fatal(err)
	}

	outputDir := fmt.Sprintf("%s/nginx", cfg.SharePath)
	_ = os.Mkdir(outputDir, os.ModePerm)

	if err := makeNginxConfig(ctx, dapps, outputDir); err != nil {
		logger.Fatal(err)
	}

	// if err := makeSitemap(ctx, dapps, outputDir); err != nil {
	// 	logger.Fatal(err)
	// }
}
