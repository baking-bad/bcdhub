package main

import (
	"fmt"
	"os"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
)

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	ctx := config.NewContext(
		config.WithStorage(cfg.Storage),
		config.WithConfigCopy(cfg),
	)
	defer ctx.Close()

	dapps, err := ctx.TZIP.GetDApps()
	if err != nil {
		logger.Fatal(err)
	}

	aliases, err := ctx.TZIP.GetAliases(consts.Mainnet)
	if err != nil {
		logger.Fatal(err)
	}

	outputDir := fmt.Sprintf("%s/nginx", cfg.SharePath)
	_ = os.Mkdir(outputDir, os.ModePerm)

	env := os.Getenv("BCD_ENV")
	if env == "" {
		logger.Fatal(fmt.Errorf("BCD_ENV env var is empty"))
	}

	nginxConfigFilename := fmt.Sprintf("%s/default.%s.conf", outputDir, env)
	if err := makeNginxConfig(dapps, aliases, nginxConfigFilename, ctx.Config.BaseURL); err != nil {
		logger.Fatal(err)
	}

	sitemapFilename := fmt.Sprintf("%s/sitemap.%s.xml", outputDir, env)
	if err := makeSitemap(dapps, aliases, sitemapFilename, ctx.Config); err != nil {
		logger.Fatal(err)
	}
}
