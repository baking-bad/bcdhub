package main

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/scripts/nginx/pkg/sitemap"
	"github.com/rs/zerolog/log"
)

func makeSitemap(filepath string, cfg config.Config) error {
	s := sitemap.New()

	s.AddLocation(cfg.BaseURL)
	s.AddLocation(fmt.Sprintf("%s/stats", cfg.BaseURL))
	s.AddLocation(fmt.Sprintf("%s/search", cfg.BaseURL))
	s.AddLocation(fmt.Sprintf("%s/dapps/list", cfg.BaseURL))

	for _, network := range cfg.Scripts.Networks {
		s.AddLocation(fmt.Sprintf("%s/stats/%s", cfg.BaseURL, network))
	}
	if err := s.SaveToFile(filepath); err != nil {
		return err
	}

	log.Info().Msgf("Sitemap created in %s", filepath)

	return nil
}
