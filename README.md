# Better Call Dev Hub
[![Build Status](https://travis-ci.org/baking-bad/bcdhub.svg?branch=master)](https://travis-ci.org/baking-bad/bcdhub)
[![made_with golang](https://img.shields.io/badge/made_with-golang-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Backend for the Better Call Dev contract explorer & developer dashboard  
https://better-call.dev

## How to run

#### Requirements
* docker
* docker-compose

#### Clone the  repo
```bash
git clone https://github.com/baking-bad/bcdhub.git
cd bcdhub
```

#### Place `.env` file
Place env file to the project folder with the following content:
```
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
GITLAB_CLIENT_ID=
GITLAB_CLIENT_SECRET=
JWT_SECRET_KEY=
OAUTH_STATE_STRING=
BCD_AWS_BUCKET_NAME=
BCD_AWS_REGION=
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
POSTGRES_USER=
POSTGRES_PASSWORD=
POSTGRES_DB=
GIN_MODE=debug
BCD_ENV=development
```

#### Option 1. Build & run
```bash
docker-compose up -d --build
```

#### Option 2. Pull & run
Using images from the dockerhub (need to specify tag):  
https://hub.docker.com/repository/docker/bakingbad/bcdhub-api  
https://hub.docker.com/repository/docker/bakingbad/bcdhub-indexer  
https://hub.docker.com/repository/docker/bakingbad/bcdhub-opindexer  
https://hub.docker.com/repository/docker/bakingbad/bcdhub-metrics  

```bash
TAG=2.0.0 make deploy
```

## Sponsored by
[Tezos Foundation](https://tezos.foundation/)
