package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

const (
	ogTitle          = "Better Call Dev"
	ogDescription    = "Tezos smart contract explorer, developer dashboard, and API provider. Easy to spin up / integrate with your sandbox."
	ogImage          = "/img/logo_og.png"
	pageTitle        = "Better Call Dev — Tezos smart contract explorer by Baking Bad"
	dappsTitle       = "Tezos DApps | Better Call Dev"
	dappsDescription = "Track the Tezos ecosystem growth: aggregated DApps usage stats, DEX token turnover, affiliated smart contracts, screenshots, social links, and more."
)

const defaultConfTemplate = `server {
	listen          80;
	server_name     localhost;
	root            /usr/share/nginx/html;
	index           index.html index.htm;
	%s
	location / {
		try_files $uri /index.html;
	}
  
	location /v1 {
		proxy_pass http://api:14000;
		proxy_http_version 1.1;
		proxy_set_header Upgrade $http_upgrade;
		proxy_set_header Connection "upgrade";
	}
}`

const dappLocationTemplate = `
	location /dapps/{{.slug}} {
		rewrite ^ /index.html break;
		sub_filter '<meta property=og:url content=/' '<meta property=og:url content={{.baseUrl}}/dapps/{{.slug}}';
		sub_filter '<meta property=og:title content="{{.ogTitle}}"' '<meta property=og:title content="{{.name}}"';
		sub_filter '<meta property=og:description content="{{.ogDescription}}"' '<meta property=og:description content="{{.description}}"';
		sub_filter '<meta property=og:image content={{.ogImage}}' '<meta property=og:image content={{.logoURL}}';
		sub_filter '<meta property=og:image:secure_url content={{.ogImage}}' '<meta property=og:image:secure_url content={{.logoURL}}';
		sub_filter '<meta name=twitter:image content={{.ogImage}}' '<meta name=twitter:image content={{.logoURL}}';
		sub_filter '<meta name=twitter:title content="{{.ogTitle}}"' '<meta name=twitter:title content="{{.name}}"';
		sub_filter '<meta name=twitter:description content="{{.ogDescription}}"' '<meta name=twitter:description content="{{.description}}"';
		sub_filter '<title>{{.pageTitle}}</title>' '<title>{{.title}}</title>';
		sub_filter_once on;
	}`

func makeDappLocation(dapp tzip.DApp, baseURL string) (string, error) {
	logoURL := ""
	for _, picture := range dapp.Pictures {
		if picture.Type == "logo" {
			logoURL = picture.Link
		}
	}

	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("").Parse(dappLocationTemplate))

	err := tmpl.Execute(buf, map[string]interface{}{
		"slug":          dapp.Slug,
		"name":          dapp.Name,
		"description":   dapp.FullDescription,
		"ogTitle":       ogTitle,
		"ogDescription": ogDescription,
		"ogImage":       ogImage,
		"pageTitle":     pageTitle,
		"baseUrl":       baseURL,
		"logoURL":       logoURL,
		"title":         fmt.Sprintf("%s | %s", dapp.Name, dappsTitle),
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func makeDappRootLocation(path, baseURL string) (string, error) {
	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("").Parse(dappLocationTemplate))

	err := tmpl.Execute(buf, map[string]interface{}{
		"slug":          path,
		"name":          dappsTitle,
		"description":   dappsDescription,
		"ogTitle":       ogTitle,
		"ogDescription": ogDescription,
		"ogImage":       ogImage,
		"pageTitle":     pageTitle,
		"baseUrl":       baseURL,
		"logoURL":       ogImage,
		"title":         dappsTitle,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func makeNginxConfig(ctx *config.Context, dapps []tzip.DApp, outputDir string) error {
	filePath := fmt.Sprintf("%s/default.conf", outputDir)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var dappLocations strings.Builder
	for _, dapp := range dapps {
		loc, err := makeDappLocation(dapp, ctx.Config.BaseURL)
		if err != nil {
			return err
		}
		dappLocations.WriteString(loc)
		dappLocations.WriteString("\n")
	}

	for _, path := range []string{"", "list"} {
		loc, err := makeDappRootLocation(path, ctx.Config.BaseURL)
		if err != nil {
			return err
		}
		dappLocations.WriteString(loc)
		dappLocations.WriteString("\n")
	}

	defaultConf := fmt.Sprintf(defaultConfTemplate, dappLocations.String())
	if _, err = file.WriteString(defaultConf); err != nil {
		logger.Fatal(err)
	}

	logger.Info("nginx default config created in %s", filePath)

	return nil
}
