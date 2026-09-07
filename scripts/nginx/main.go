package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/rs/zerolog/log"
)

func main() {
	cctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		log.Err(err).Msg("load config")
		return
	}

	ctx := config.NewContext(
		types.Mainnet,
		config.WithStorage(cctx, cfg.Storage, "nginx", 0, time.Second*10),
		config.WithConfigCopy(cfg),
	)
	defer ctx.Close()

	outputDir := filepath.Join(cfg.SharePath, "nginx")
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		log.Err(err).Msg("mkdir")
		return
	}

	env := os.Getenv("BCD_ENV")
	if env == "" {
		log.Error().Msg("BCD_ENV env var is empty")
		return
	}

	nginxConfigFilename := fmt.Sprintf("%s/default.%s.conf", outputDir, env)
	if err := makeNginxConfig(nginxConfigFilename); err != nil {
		log.Err(err).Msg("make nginx config")
		return
	}

	sitemapFilename := fmt.Sprintf("%s/sitemap.%s.xml", outputDir, env)
	if err := makeSitemap(sitemapFilename, ctx.Config); err != nil {
		log.Err(err).Msg("make sitemap")
		return
	}
}
