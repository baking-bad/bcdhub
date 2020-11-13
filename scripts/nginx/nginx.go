package main

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"os"
	"strings"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

const ogTitle = "Better Call Dev"
const ogDescription = "Tezos smart contract explorer, developer dashboard, and API provider. Easy to spin up / integrate with your sandbox."
const ogImage = "/img/logo_og.png"
const pageTitle = "Better Call Dev — Tezos smart contract explorer by Baking Bad"

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
		sub_filter '<meta property=og:title content="{{.ogTitle}}"' '<meta property=og:title content="{{.name}} | Tezos DApps"';
		sub_filter '<meta property=og:description content="{{.ogDescription}}"' '<meta property=og:description content="{{.shortDescription}}"';
		sub_filter '<meta property=og:image content={{.ogImage}}' '<meta property=og:image content={{.logoURL}}';
		sub_filter '<meta property=og:image:secure_url content={{.ogImage}}' '<meta property=og:image:secure_url content={{.logoURL}}';
		sub_filter '<meta name=twitter:image content={{.ogImage}}' '<meta name=twitter:image content={{.logoURL}}';
		sub_filter '<meta name=twitter:title content="{{.ogTitle}}"' '<meta name=twitter:title content="{{.name}}"';
		sub_filter '<meta name=twitter:description content="{{.ogDescription}}"' '<meta name=twitter:description content="{{.shortDescription}}"';
		sub_filter '<title>{{.pageTitle}}</title>' '<title>{{.name}} | Tezos DApps</title>';
		sub_filter_once on;
	}`

func makeDappLocation(dapp tzip.DApp, baseURL string) string {
	logoURL := ""
	for _, picture := range dapp.Pictures {
		if picture.Type == "logo" {
			logoURL = picture.Link
		}
	}

	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("").Parse(html.EscapeString(dappLocationTemplate)))

	err := tmpl.Execute(buf, map[string]interface{}{
		"slug":             dapp.Slug,
		"name":             dapp.Name,
		"shortDescription": dapp.ShortDescription,
		"ogTitle":          ogTitle,
		"ogDescription":    ogDescription,
		"ogImage":          ogImage,
		"pageTitle":        pageTitle,
		"baseUrl":          baseURL,
		"logoURL":          logoURL,
	})
	if err != nil {
		logger.Fatal(err)
	}

	return html.UnescapeString(html.UnescapeString(buf.String()))
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
		dappLocations.WriteString(makeDappLocation(dapp, ctx.Config.BaseURL))
		dappLocations.WriteString("\n")
	}

	defaultConf := fmt.Sprintf(defaultConfTemplate, dappLocations.String())
	if _, err = file.WriteString(defaultConf); err != nil {
		logger.Fatal(err)
	}

	logger.Info("nginx default config created in %s", filePath)

	return nil
}
