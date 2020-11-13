package main

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/scripts/nginx/pkg/sitemap"
)

func makeSitemap(ctx *config.Context, dapps []tzip.DApp, outputDir string) error {
	aliases, err := ctx.ES.GetAliases(consts.Mainnet)
	if err != nil {
		return err
	}

	aliasModels := make([]models.TZIP, 0)

	for _, a := range aliases {
		if a.Slug == "" {
			continue
		}

		data := models.TZIP{
			Address: a.Address,
			Network: a.Network,
		}
		if err := ctx.ES.GetByID(&data); err != nil {
			continue
		}

		logger.Info("%s %s", a.Address, data.Name)

		aliasModels = append(aliasModels, a)
	}

	// logger.Info("Total aliases: %d", len(aliasModels))

	if err := buildXML(aliasModels, ctx.Config, dapps, outputDir); err != nil {
		return err
	}

	logger.Info("Sitemap created in sitemap.xml")

	return nil
}

func buildXML(aliases []models.TZIP, cfg config.Config, dapps []tzip.DApp, outputDir string) error {
	s := sitemap.New()

	s.AddLocation(cfg.BaseURL)
	s.AddLocation(fmt.Sprintf("%s/stats", cfg.BaseURL))
	s.AddLocation(fmt.Sprintf("%s/search", cfg.BaseURL))
	s.AddLocation(fmt.Sprintf("%s/dapps/list", cfg.BaseURL))
	s.AddLocation(fmt.Sprintf("https://%s/v1/docs/index.html", cfg.API.SwaggerHost))

	for _, network := range cfg.Scripts.Networks {
		s.AddLocation(fmt.Sprintf("%s/stats/%s", cfg.BaseURL, network))
	}

	for _, a := range aliases {
		s.AddLocation(fmt.Sprintf("%s/@%s", cfg.BaseURL, a.Slug))
	}

	for _, d := range dapps {
		s.AddLocation(fmt.Sprintf("%s/dapps/%s", cfg.BaseURL, d.Slug))
	}

	return s.SaveToFile(fmt.Sprintf("%s/sitemap.xml", outputDir))
}
