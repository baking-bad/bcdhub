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

func makeSitemap(ctx *config.Context, dapps []tzip.DApp) error {
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

		// logger.Info("%s %s", a.Address, data.Name)

		aliasModels = append(aliasModels, a)
	}

	// logger.Info("Total aliases: %d", len(aliasModels))

	if err := buildXML(aliasModels, ctx.Config.Scripts.Networks, dapps); err != nil {
		return err
	}

	logger.Info("Sitemap created in sitemap.xml")

	return nil
}

func buildXML(aliases []models.TZIP, networks []string, dapps []tzip.DApp) error {
	s := sitemap.New()

	s.AddLocation("https://better-call.dev")
	s.AddLocation("https://better-call.dev/stats")
	s.AddLocation("https://better-call.dev/search")
	s.AddLocation("https://better-call.dev/dapps/list")
	s.AddLocation("https://api.better-call.dev/v1/docs/index.html")

	for _, network := range networks {
		s.AddLocation(fmt.Sprintf("https://better-call.dev/stats/%s", network))
	}

	for _, a := range aliases {
		s.AddLocation(fmt.Sprintf("https://better-call.dev/@%s", a.Slug))
	}

	for _, d := range dapps {
		s.AddLocation(fmt.Sprintf("https://better-call.dev/dapps/%s", d.Slug))
	}

	return s.SaveToFile("../../build/nginx/sitemap.xml")
}
