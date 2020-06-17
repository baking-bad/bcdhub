# BCD Hub
[![Build Status](https://travis-ci.org/baking-bad/bcdhub.svg?branch=master)](https://travis-ci.org/baking-bad/bcdhub)
[![Docker Build Status](https://img.shields.io/docker/cloud/build/bakingbad/bcdhub-api)](https://hub.docker.com/r/bakingbad/bcdhub-api)
[![made_with golang](https://img.shields.io/badge/made_with-golang-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Backend for the [Better Call Dev](https://better-call.dev) contract explorer & developer dashboard.

## Overview
BCDHub is a set of microservices written in Golang:

* `Indexer`  
Loads and decodes operations related to smart contracts and also keeps track of the block chain and handles protocol updates.    
* `Metrics`  
Receives new contract/operation events from the indexer and calculates various metrics that are used for ranking, linking, and labelling contracts and operations.
* `API`  
Exposes RESTful JSON API for accessing indexed data (with on-the-fly decoding). Also provides a set of methods for authentication and managing user profiles.

Those microservices are sharing access to databases and communicating via message queue:

* `ElasticSearch` cluster (single node) for storing all indexed data including blocks, protocols, contracts, operations, Big_map diffs, and others.
* `PostgreSQL` database for storing off-chain data such as contract aliases and user data.
* `RabbitMQ` for one-way communications `indexer` -> `metrics`, `API`.

### Third-party services
BCDHub also depends on several API endpoints exposed by [TzKT](https://github.com/baking-bad/tzkt) although they are optional:

* List of blocks containing smart contract operations, used for boosting the indexing process (allows to skip blocks with no contract calls)
* Mempool operations
* Contract aliases and other metadata

Those services are obviously make sense for public networks only and not used for sandbox or other private environments.

## Versioning
BCD uses `X.Y.Z` version format where:
* `X` changes every 3-5 months along with a big release with a significant addition of functionality  
* `Y` increasing signals about a possibly non-compatible update that requires reindexing (or restoring from snaphot) or syncing with frontend
* `Z` bumped for every stable release candidate or hotfix

### Syncing with frontend
BCD web interface developed at https://github.com/baking-bad/bcd uses the same version scheme.  
`X.Y.*` versions of backend and frontent MUST BE compatible which means that for every change in API responses `Y` has to be increased.

### Publishing releases
Is essentially tagging commits:
```bash
make latest  # forced tag update
```
For stable release:
```bash
git tag X.Y.Z
git push --tags
```

## Docker images
Although you can install and run each part of BCD Hub independently, as system services for instance, the simplest approach is to use dockerized versions orchestrated by _docker-compose_.  

BCDHub docker images are being built on [dockerhub](https://hub.docker.com/u/bakingbad). Two types of tags are provided:
* `latest` should be considered experimental
* `X.Y` stable releases

### Linking with Git tags
Docker tags are essentially produced from Git tags using the following rules:
* `latest` → `latest`
* `X.Y.*` → `X.Y`

### Building images
```bash
make images  # latest
make stable-images  # requires STABLE_TAG variable in the .env file
```

## Configuration
BCD configuration is stored in _yml_ files in docker-compose style: you can **merge** multiple configs and **expand** environment variables.  

Each service has its very own section in the config file and also they share several common sections. There are predefined configs for _dev_, _prod_, and _sandbox_ environments.

### Main config `config.yml`

#### `rpc`
List of RPC nodes with base urls and connection timeouts
```yml
rpc:
    mainnet:
        uri: https://mainnet-tezos.giganode.io
        timeout: 20
```

#### `tzkt`
TzKT API endpoints (optional) and connection timeouts
```yml
tzkt:
    mainnet:
        uri: https://api.tzkt.io/v1/
        services_uri: https://services.tzkt.io/v1/
        timeout: 20
```

#### `share`
Folder to store cached contract sources
```yml
share:
    path: /etc/bcd
```

#### `sentry`
[Sentry](https://sentry.io/) configuration
```yml
sentry:
    environment: production
    uri: ${SENTRY_DSN}
    debug: false
```

#### `aws`
[AWS S3](https://aws.amazon.com/s3/) snapshot registry settings
```yml
aws:
    bucket_name: bcd-elastic-snapshots
    region: eu-central-1
```

#### `oauth`
OAuth providers settings
```yml
oauth:
    state: ${OAUTH_STATE_STRING}
    jwt:
        secret: ${JWT_SECRET_KEY}
        redirect_url: https://better-call.dev/welcome
    github:
        client_id: ${GITHUB_CLIENT_ID}
        secret: ${GITHUB_CLIENT_SECRET}
        callback_url: https://api.better-call.dev/v1/oauth/github/callback
    gitlab:
        client_id: ${GITLAB_CLIENT_ID}
        secret: ${GITLAB_CLIENT_SECRET}
        callback_url: https://api.better-call.dev/v1/oauth/gitlab/callback

```

#### `elastic`
Elastic Search configuration
```yml
elastic:
    uri: http://elastic:9200
    timeout: 10
```

#### `db`
PostgreSQL connection string
```yml
db:
    conn_string: "host=db port=5432 user=${POSTGRES_USER} dbname=bcd password=${POSTGRES_PASSWORD} sslmode=disable"
```

#### `rabbitmq`
RabbitMQ settings and list of queues to subscribe
```yml
rabbitmq:
    uri: "amqp://${RABBITMQ_DEFAULT_USER}:${RABBITMQ_DEFAULT_PASS}@mq:5672/"
    queues:
        - operations
        - recalc
```

#### `seed`
Prepopulated data: default user (for sandbox mode), user datam and contract aliases
```yml
seed:
    user:
        login: sandboxuser
        name: "Default User"
        avatar_url: "https://services.tzkt.io/v1/avatars/bcd"
        token: ""
    subscriptions:
        - address: tz1VSUr8wwNhLAzempoch5d6hLRiTh8Cjcjb
          network: sandboxnet
          alias: Alice
          watch_mask: 127
    aliases:
        - alias: Alice
          network: sandboxnet
          address: tz1VSUr8wwNhLAzempoch5d6hLRiTh8Cjcjb
```

#### `api`
API service settings: 
```yml
api:
    bind: ":14000"
    swagger_host: "api.better-call.dev"
    oauth:
        enabled: true
    sentry:
        enabled: true
        project: api
    networks:
        - mainnet
    seed:
        enabled: false
```

#### `indexer`
Indexer service settings. Note the optional _boost_ setting which tells indexer to use third-party service in order to speed up the process.
```yml
indexer:
    sentry:
        enabled: true
        project: indexer
    networks:
        mainnet:
          boost: tzkt
```

#### `metrics`
Metrics service settings
```yml
metrics:
    sentry:
        enabled: true
        project: metrics
```

### Docker settings `docker-compose.yml`
Connects all the services together. The compose file is pretty straightforward and universal, although there are several settings you may want to change:

* Container names
* Ports
* Shared paths

If you are altering these settings make sure you are in sync with the `config.yml` (also keep in mind that container names are essentially hostnames in the internal docker network).

#### Local RPC node
A typical problem is to access service running on the host machine from inside a docker container. Currently there's no unversal (cross-platform) way to do it (should be fixed in docker 20). A suggested way is the following:

1. Expose your node at `172.17.0.1:8732` (docker gateway)
2. For each docker service that needs to access RPC add to compose file:
    ```yml
    extra_hosts:
        sandbox: 172.17.0.1
    ```
3. Now you can update configuration:
    ```yml
    rpc:
        sandboxnet:
            uri: http://sandbox:8732
            timeout: 20     
    ```

### Environment variables `.env`
About env files: https://docs.docker.com/compose/env-file/

#### System config _required_
* `GIN_MODE` _release_ for production, _debug_ otherwise
* `ES_JAVA_OPTS` _"-Xms1g -Xmx1g"_ max RAM allocation for Elastic Search (_g_ for GB, _m_ for MB)

#### Credentials _required_
* `POSTGRES_USER` e.g. _root_
* `POSTGRES_PASSWORD` e.g. _root_
* `POSTGRES_DB` e.g. _bcd_
* `RABBITMQ_DEFAULT_USER` e.g. _guest_
* `RABBITMQ_DEFAULT_PASS` e.g. _guest_

#### OAuth creds _required if `oauth: enabled`_
* `GITHUB_CLIENT_ID`
* `GITHUB_CLIENT_SECRET`
* `GITLAB_CLIENT_ID`
* `GITLAB_CLIENT_SECRET`
* `JWT_SECRET_KEY`
* `OAUTH_STATE_STRING`

#### Sentry creds _required if `sentry: enabled`_
* `SENTRY_DSN`

#### Snapshot settings
* `BCD_AWS_BUCKET_NAME`
* `BCD_AWS_REGION`
* `AWS_ACCESS_KEY_ID`
* `AWS_SECRET_ACCESS_KEY`

#### Others
* `STABLE_TAG` _required for building & running images_ e.g. _2.5_
* `USER_ID` _required for single-user mode_ e.g. _1_

## Deploy
If you are looking for a full-fledged BCD setup with GUI (e.g. for local development env) check out https://github.com/baking-bad/bbbox

### Requirements
Make sure you have installed:
* docker
* docker-compose

You will also need several ports to be not busy:
* `14000` API service
* `9200, 9300` Elastic
* `5672, 15672` RabbitMQ
* `5432` PostgreSQL

### Get ready
1. Clone this repo
```bash
git clone https://github.com/baking-bad/bcdhub.git
cd bcdhub
```

2. Create and fill `.env` file (see _Configuration_)
```bash
your-text-editor .env
```

### Environments
There are several predefined configurations serving different purposes.

#### Production `better-call.dev`
* Stable docker images `X.Y`
* `/cmd/{service}/config.yml` files are used internally
* Requires `STABLE_TAG` environment set
* Deployed via `make stable`

#### Staging `you.better-call.dev`
* Latest docker images `latest`
* Single `config.yml` file mapped through docker volumes
* Deployed via https://github.com/baking-bad/bbbox using `make custom`

#### Development `localhost`
* `config.yml` + `config.dev.yml` files are used (merged)
* You can spawn local instances of databases and message queue or _ssh_ to staging host with port forwarding
* Run services `make {service}` (where service is one of `api` `indexer` `metrics`)

#### Sandbox `bbbox`
See https://github.com/baking-bad/bbbox


## Running

### Startup
It takes around 20-30 seconds to initialize all services, API endpoints might return errors until then.  
**NOTE** that if you specified local RPC node that's not running, BCDHub will wait for it indefinitely.

### Single user mode
If `USER_ID` variable is set all API endpoints that are hidden behind auth become accessible without JWT token.  
This is done specifically for sandbox environment.


## Contract aliases
In order to 
```bash
make aliases
```

```bash
make migration
```

```bash
docker restart bcd-api
```

## Snapshots


## Upgrade

### Soft update

### Migrations

### Upgrade from snapshot


## About
This project is the successor of the first [serverless](https://github.com/baking-bad/better-call-dev) version (aka BCD1). It has been rewritten from scratch in Golang.   
Better Call Dev was initially funded and is currently supported by [Tezos Foundation](https://tezos.foundation/).
