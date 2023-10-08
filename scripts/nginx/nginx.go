package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
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

func makeNginxConfig(filepath, baseURL string) error {
	var locations strings.Builder

	defaultConf := fmt.Sprintf(defaultConfTemplate, locations.String())
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.WriteString(defaultConf); err != nil {
		return err
	}

	log.Info().Msgf("Nginx default config created in %s", filepath)

	return nil
}
