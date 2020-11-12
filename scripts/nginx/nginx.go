package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

var configStart = `server {
	listen          80;
	server_name     localhost;
	root            /usr/share/nginx/html;
	index           index.html index.htm;

`

var configEnd = `  location / {
	  try_files $uri /index.html;
  }

  location /v1 {
	  proxy_pass http://api:14000;
	  proxy_http_version 1.1;
	  proxy_set_header Upgrade $http_upgrade;
	  proxy_set_header Connection "upgrade";
  }
}`

func makeNginxConfig(ctx *config.Context, dapps []tzip.DApp) error {
	env := os.Getenv(config.EnvironmentVar)
	if env == "" {
		return fmt.Errorf("no %s env var", config.EnvironmentVar)
	}

	file, err := os.Create(fmt.Sprintf("../../build/nginx/%s.conf", env))
	if err != nil {
		return err
	}
	defer file.Close()

	nginxConf := buildConfig(dapps, ctx.Config.BaseURL)

	if _, err := file.WriteString(nginxConf); err != nil {
		return err
	}

	logger.Info("nginx default config created in build/nginx/%s.xml", env)

	return nil
}

func buildConfig(dapps []tzip.DApp, baseURL string) string {
	var config strings.Builder

	config.WriteString(configStart)

	for _, dapp := range dapps {
		config.WriteString(fmt.Sprintf("  location /dapps/%s {\n", dapp.Slug))
		config.WriteString("    rewrite ^ /index.html break;\n")
		config.WriteString(fmt.Sprintf("    sub_filter '<meta property=og:url content=/' '<meta property=og:url content=%s/dapps/%s';\n", baseURL, dapp.Slug))
		config.WriteString(fmt.Sprintf("    sub_filter '<meta property=og:title content=\"Better Call Dev\"' '<meta property=og:title content=\"%s | Tezos DApps\"';\n", dapp.Name))
		config.WriteString(fmt.Sprintf("    sub_filter '<meta property=og:description content=\"Tezos smart contract explorer, developer dashboard, and API provider. Easy to spin up / integrate with your sandbox.\"' '<meta property=og:description content=\"%s\"';\n", dapp.ShortDescription))

		for _, picture := range dapp.Pictures {
			if picture.Type == "logo" {
				config.WriteString(fmt.Sprintf("    sub_filter '<meta property=og:image content=/img/logo_og.png' '<meta property=og:image content=%s';\n", picture.Link))
				config.WriteString(fmt.Sprintf("    sub_filter '<meta property=og:image:secure_url content=/img/logo_og.png' '<meta property=og:image:secure_url content=%s';\n", picture.Link))
				config.WriteString(fmt.Sprintf("    sub_filter '<meta name=twitter:image content=/img/logo_og.png' '<meta name=twitter:image content=%s';\n", picture.Link))

				break
			}
		}

		config.WriteString(fmt.Sprintf("    sub_filter '<meta name=twitter:title content=\"Better Call Dev\"' '<meta name=twitter:title content=\"%s\"';\n", dapp.Name))
		config.WriteString(fmt.Sprintf("    sub_filter '<meta name=twitter:description content=\"Tezos smart contract explorer, developer dashboard, and API provider. Easy to spin up / integrate with your sandbox.\"' '<meta name=twitter:description content=\"%s\"';\n", dapp.ShortDescription))
		config.WriteString(fmt.Sprintf("    sub_filter '<title>Better Call Dev — Tezos smart contract explorer by Baking Bad</title>' '<title>%s | Tezos DApps</title>';\n", dapp.Name))
		config.WriteString("    sub_filter_once on;\n  }\n\n")
	}

	config.WriteString(configEnd)

	return config.String()
}
