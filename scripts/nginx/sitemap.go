package main

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/scripts/nginx/pkg/sitemap"
)

func makeSitemap(aliases []contract_metadata.ContractMetadata, filepath string, cfg config.Config) error {
	s := sitemap.New()

	s.AddLocation(cfg.BaseURL)
	s.AddLocation(fmt.Sprintf("%s/stats", cfg.BaseURL))
	s.AddLocation(fmt.Sprintf("%s/search", cfg.BaseURL))
	s.AddLocation(fmt.Sprintf("https://%s/v1/docs/index.html", cfg.API.SwaggerHost))

	for _, network := range cfg.Scripts.Networks {
		s.AddLocation(fmt.Sprintf("%s/stats/%s", cfg.BaseURL, network))
	}

	for _, a := range aliases {
		s.AddLocation(fmt.Sprintf("%s/@%s", cfg.BaseURL, a.Slug))
	}

	if err := s.SaveToFile(filepath); err != nil {
		return err
	}

	logger.Info().Msgf("Sitemap created in %s", filepath)

	return nil
}
