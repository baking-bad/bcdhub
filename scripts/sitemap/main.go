package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/logger"
)

// URL is awesome
type URL struct {
	XMLName  xml.Name `xml:"url"`
	Location string   `xml:"loc"`
	LastMod  string   `xml:"lastmod"`
}

// URLSet is awesome
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

func buildXML(aliases []database.Alias, networks []string) error {
	u := &URLSet{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	modDate := time.Now().Format("2006-01-02")

	u.URLs = append(u.URLs, URL{Location: "https://better-call.dev", LastMod: modDate})
	u.URLs = append(u.URLs, URL{Location: "https://better-call.dev/stats", LastMod: modDate})
	u.URLs = append(u.URLs, URL{Location: "https://better-call.dev/search", LastMod: modDate})
	u.URLs = append(u.URLs, URL{Location: "https://api.better-call.dev/v1/docs/index.html", LastMod: modDate})

	for _, network := range networks {
		loc := fmt.Sprintf("https://better-call.dev/stats/%s", network)
		u.URLs = append(u.URLs, URL{Location: loc, LastMod: modDate})
	}

	for _, a := range aliases {
		loc := fmt.Sprintf("https://better-call.dev/@%s", a.Slug)
		u.URLs = append(u.URLs, URL{Location: loc, LastMod: modDate})
	}

	file, err := os.Create("sitemap.xml")
	if err != nil {
		return err
	}

	xmlWriter := io.Writer(file)

	if _, err = xmlWriter.Write([]byte(xml.Header)); err != nil {
		return err
	}

	enc := xml.NewEncoder(xmlWriter)
	enc.Indent("", "  ")
	if err := enc.Encode(u); err != nil {
		return fmt.Errorf("encode error: %v", err)
	}

	return nil
}

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		logger.Fatal(err)
	}

	ctx := config.NewContext(
		config.WithElasticSearch(cfg.Elastic),
		config.WithDatabase(cfg.DB),
	)
	defer ctx.Close()

	aliases, err := ctx.DB.GetAliases(consts.Mainnet)
	if err != nil {
		logger.Fatal(err)
	}

	var contracts []database.Alias

	for _, a := range aliases {
		if strings.HasPrefix(a.Address, "tz") || a.Slug == "" {
			continue
		}

		by := map[string]interface{}{
			"address": a.Address,
			"network": a.Network,
		}
		cntr, err := ctx.ES.GetContract(by)
		if err != nil {
			continue
		}

		logger.Info("%s %s", a.Address, cntr.Alias)

		contracts = append(contracts, a)
	}

	logger.Info("Total contract aliases: %d", len(contracts))

	if err := buildXML(contracts, cfg.Migrations.Networks); err != nil {
		logger.Fatal(err)
	}

	logger.Success("Sitemap created in sitemap.xml")
}
