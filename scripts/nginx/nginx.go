package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

const (
	ogTitle             = "Better Call Dev"
	ogDescription       = "Tezos smart contract explorer, developer dashboard, and API provider. Easy to spin up / integrate with your sandbox."
	ogImage             = "/img/logo_og.png"
	pageTitle           = "Better Call Dev — Tezos smart contract explorer by Baking Bad"
	dappsTitle          = "Tezos DApps"
	pageDescription     = "Tezos smart contract explorer & developer dashboard, simplifies perception and facilitates interaction. By Baking Bad."
	dappsDescription    = "Track the Tezos ecosystem growth: aggregated DApps usage stats, DEX token turnover, affiliated smart contracts, screenshots, social links, and more."
	contractDescription = "Check out recent operations, inspect contract code and storage, invoke contract methods."
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

const locationTemplate = `
	location {{.location}} {
		rewrite ^ /index.html break;
		sub_filter '<meta property=og:url content=/' '<meta property=og:url content={{.url}}';
		sub_filter '<meta property=og:title content="{{.ogTitle}}"' '<meta property=og:title content="{{.title}}"';
		sub_filter '<meta property=og:description content="{{.ogDescription}}"' '<meta property=og:description content="{{.description}}"';
		sub_filter '<meta property=og:image content={{.ogImage}}' '<meta property=og:image content={{.logoURL}}';
		sub_filter '<meta property=og:image:secure_url content={{.ogImage}}' '<meta property=og:image:secure_url content={{.logoURL}}';
		sub_filter '<meta name=twitter:image content={{.ogImage}}' '<meta name=twitter:image content={{.logoURL}}';
		sub_filter '<meta name=twitter:title content="{{.ogTitle}}"' '<meta name=twitter:title content="{{.title}}"';
		sub_filter '<meta name=twitter:description content="{{.ogDescription}}"' '<meta name=twitter:description content="{{.description}}"';
		sub_filter '<meta name=description content="{{.pageDescription}}"' '<meta name=description content="{{.description}}"';
		sub_filter '<title>{{.pageTitle}}</title>' '<title>{{.title}}</title>';
		sub_filter_once on;
	}`

func makeNginxConfig(dapps []tzip.DApp, aliases []tzip.TZIP, filepath, baseURL string) error {
	var locations strings.Builder
	tmpl := template.Must(template.New("").Parse(locationTemplate))

	for _, dapp := range dapps {
		loc, err := makeDappLocation(tmpl, dapp, baseURL)
		if err != nil {
			return err
		}
		locations.WriteString(loc)
		locations.WriteString("\n")
	}

	loc, err := makeDappRootLocation(tmpl, "list", baseURL)
	if err != nil {
		return err
	}
	locations.WriteString(loc)
	locations.WriteString("\n")

	for _, alias := range aliases {
		loc, err := makeContractsLocation(tmpl, alias.Address, alias.Name, baseURL)
		if err != nil {
			return err
		}

		locations.WriteString(loc)
		locations.WriteString("\n")
	}

	defaultConf := fmt.Sprintf(defaultConfTemplate, locations.String())
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.WriteString(defaultConf); err != nil {
		logger.Fatal(err)
	}

	logger.Info("Nginx default config created in %s", filepath)

	return nil
}

func makeDappLocation(tmpl *template.Template, dapp tzip.DApp, baseURL string) (string, error) {
	var logoURL string
	for _, picture := range dapp.Pictures {
		if picture.Type == "logo" {
			logoURL = picture.Link
		}
	}

	buf := new(bytes.Buffer)
	err := tmpl.Execute(buf, map[string]interface{}{
		"location":        fmt.Sprintf("/dapps/%s", sanitizeQuotes(dapp.Slug)),
		"url":             fmt.Sprintf("%s/dapps/%s", baseURL, sanitizeQuotes(dapp.Slug)),
		"title":           fmt.Sprintf("%s — %s", sanitizeQuotes(dapp.Name), sanitizeQuotes(dapp.ShortDescription)),
		"description":     sanitizeQuotes(dapp.FullDescription),
		"ogTitle":         ogTitle,
		"ogDescription":   ogDescription,
		"ogImage":         ogImage,
		"pageTitle":       pageTitle,
		"pageDescription": pageDescription,
		"logoURL":         logoURL,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func makeDappRootLocation(tmpl *template.Template, path, baseURL string) (string, error) {
	buf := new(bytes.Buffer)
	err := tmpl.Execute(buf, map[string]interface{}{
		"location":        fmt.Sprintf("/dapps/%s", path),
		"url":             fmt.Sprintf("%s/dapps/%s", baseURL, path),
		"title":           dappsTitle,
		"description":     dappsDescription,
		"ogTitle":         ogTitle,
		"ogDescription":   ogDescription,
		"ogImage":         ogImage,
		"pageTitle":       pageTitle,
		"pageDescription": pageDescription,
		"logoURL":         ogImage,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func makeContractsLocation(tmpl *template.Template, address, alias, baseURL string) (string, error) {
	buf := new(bytes.Buffer)
	err := tmpl.Execute(buf, map[string]interface{}{
		"location":        fmt.Sprintf("/mainnet/%s", address),
		"url":             fmt.Sprintf("%s/mainnet/%s", baseURL, address),
		"title":           fmt.Sprintf("%s — %s", sanitizeQuotes(alias), ogTitle),
		"description":     contractDescription,
		"ogTitle":         ogTitle,
		"ogDescription":   ogDescription,
		"ogImage":         ogImage,
		"pageTitle":       pageTitle,
		"pageDescription": pageDescription,
		"logoURL":         ogImage,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func sanitizeQuotes(str string) string {
	return strings.ReplaceAll(str, "'", "’")
}
