# BCD Hub
[![Build Status](https://github.com/baking-bad/bcdhub/workflows/build/badge.svg)](https://github.com/baking-bad/bcdhub/actions?query=branch%3Amaster+workflow%3A%22build%22)
[![Docker Build Status](https://img.shields.io/docker/cloud/build/bakingbad/bcdhub-api)](https://hub.docker.com/r/bakingbad/bcdhub-api)
[![made_with golang](https://img.shields.io/badge/made_with-golang-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Backend for the [Better Call Dev](https://better-call.dev) contract explorer & developer dashboard.

## Overview
BCDHub is a set of microservices written in Golang:

* `indexer`  
Loads and decodes operations related to smart contracts and also keeps track of the block chain and handles protocol updates.    
* `metrics`  
Receives new contract/operation events from the indexer and calculates various metrics that are used for ranking, linking, and labelling contracts and operations.
* `API`  
Exposes RESTful JSON API for accessing indexed data (with on-the-fly decoding). Also provides a set of methods for authentication and managing user profiles.
* `compiler`  
Contains compilers of various high-level contract languages (LIGO, SmartPy, etc) as well as a service handling compilation tasks

Those microservices are sharing access to databases and communicating via message queue:

* `ElasticSearch` cluster (single node) for storing all indexed data including blocks, protocols, contracts, operations, Big_map diffs, and others.
* `PostgreSQL` database for storing compilations and user data.
* `RabbitMQ` for communications between `API`, `indexer`, `metrics` and `compiler`.

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
make release  # forced tag update
```
For stable release:
```bash
git tag X.Y.Z
git push --tags
```

## Docker images
Although you can install and run each part of BCD Hub independently, as system services for instance, the simplest approach is to use dockerized versions orchestrated by _docker-compose_.  

BCDHub docker images are being built on [dockerhub](https://hub.docker.com/u/bakingbad). Tags for stable releases have format `X.Y`.

### Linking with Git tags
Docker tags are essentially produced from Git tags using the following rules:
* `X.Y.*` → `X.Y`

### Building images
```bash
make images  # latest
make stable-images  # requires STABLE_TAG variable in the .env file
```

## Configuration
BCD configuration is stored in _yml_ files: you can **expand** environment variables.  

Each service has its very own section in the config file and also they share several common sections. There are predefined configs for _production_, _development_, _sandbox_ and _staging_ environments.

### Production config `./configs/production.yml`

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
        base_uri: https://tzkt.io/
        timeout: 20
```

#### `elastic`
Elastic Search configuration
```yml
elastic:
    uri:
        - http://elastic:9200
    timeout: 10
```

#### `rabbitmq`
RabbitMQ settings and list of queues to subscribe
```yml
rabbitmq:
    uri: "amqp://${RABBITMQ_DEFAULT_USER}:${RABBITMQ_DEFAULT_PASS}@mq:5672/"
    publisher: true
```

#### `db`
PostgreSQL connection string
```yml
db:
    conn_string: "host=db port=5432 user=${POSTGRES_USER} dbname=bcd password=${POSTGRES_PASSWORD} sslmode=disable"
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

#### `sentry`
[Sentry](https://sentry.io/) configuration
```yml
sentry:
    environment: production
    uri: ${SENTRY_DSN}
    debug: false
```

#### `share_path`
Folder to store cached contract sources and share files for `compiler`
```yml
share_path: /etc/bcd
```

#### `ipfs`
IPFS settings (list of http gateways)
```yml
ipfs:
    - https://ipfs.io
    - https://dweb.link
```

#### `api`
API service settings
```yml
api:
    project_name: api
    bind: ":14000"
    swagger_host: "api.better-call.dev"
    cors_enabled: false
    oauth_enabled: true
    sentry_enabled: true
    seed_enabled: false
    networks:
        - mainnet
    mq:
        publisher: true
        queues:
            operations:
                non_durable: true
                auto_deleted: true
```

#### `compiler`
Compiler service settings
```yml
compiler:
    project_name: compiler
    aws:
        bucket_name: bcd-contract-sources
        region: eu-central-1
        access_key_id: ${AWS_ACCESS_KEY_ID}
        secret_access_key: ${AWS_SECRET_ACCESS_KEY}
    sentry_enabled: true
    mq:
        publisher: true
        queues:
            compilations:
```

#### `indexer`
Indexer service settings. Note the optional _boost_ setting which tells indexer to use third-party service in order to speed up the process.
```yml
indexer:
    project_name: indexer
    sentry_enabled: true
    skip_delegator_blocks: false
    mq:
        publisher: true
    networks:
        mainnet:
          boost: tzkt
```

#### `metrics`
Metrics service settings
```yml
metrics:
    project_name: metrics
    sentry_enabled: true
    mq:
        publisher: false
        queues:
            operations:
            contracts:
            migrations:
            recalc:
            transfers:
            bigmapdiffs:
```

#### `scripts`
Scripts settings for data migrations and [AWS S3](https://aws.amazon.com/s3/) snapshot registry
```yml
scripts:
    aws:
        bucket_name: bcd-elastic-snapshots
        region: eu-central-1
        access_key_id: ${AWS_ACCESS_KEY_ID}
        secret_access_key: ${AWS_SECRET_ACCESS_KEY}
    networks:
      - mainnet
      - carthagenet
      - edo2net
      - florencenet
    mq:
        publisher: true
```

### Docker settings `docker-compose.yml`
Connects all the services together. The compose file is pretty straightforward and universal, although there are several settings you may want to change:

* Container names
* Ports
* Shared paths

If you are altering these settings make sure you are in sync with your `.yml` configuration file.

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
* `BCD_ENV` e.g. _production_ or _sandbox_
* `COMPOSE_PROJECT_NAME` e.g. _bcd-prod_ or _bcd-box_
* `GIN_MODE` _release_ for production, _debug_ otherwise
* `ES_JAVA_OPTS` _"-Xms1g -Xmx1g"_ max RAM allocation for Elastic Search (_g_ for GB, _m_ for MB)

#### Credentials _required_
* `POSTGRES_USER` e.g. _root_
* `POSTGRES_PASSWORD` e.g. _root_
* `POSTGRES_DB` e.g. _bcd_
* `RABBITMQ_DEFAULT_USER` e.g. _guest_
* `RABBITMQ_DEFAULT_PASS` e.g. _guest_

#### Services ports _required_
* `BCD_API_PORT` e.g. _14000_
* `ES_REQUESTS_PORT` e.g. _9200_
* `RABBITMQ_PORT` e.g. _5672_
* `POSTGRES_PORT` e.g. _5432_
* `BCD_GUI_PORT` e.g. _8000_

#### OAuth creds _required if `oauth_enabled: true`_
* `GITHUB_CLIENT_ID`
* `GITHUB_CLIENT_SECRET`
* `GITLAB_CLIENT_ID`
* `GITLAB_CLIENT_SECRET`
* `JWT_SECRET_KEY`
* `OAUTH_STATE_STRING`

#### Sentry creds _required if `sentry_enabled: true`_
* `SENTRY_DSN`

#### AWS settings
* `AWS_ACCESS_KEY_ID`
* `AWS_SECRET_ACCESS_KEY`

#### Others
* `STABLE_TAG` _required for building & running images_ e.g. _2.5_

## Deploy

### Requirements
Make sure you have installed:
* docker
* docker-compose

You will also need several ports to be not busy:
* `14000` API service
* `9200` Elastic
* `5672` RabbitMQ
* `5432` PostgreSQL
* `8000` Frontend GUI

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
* `/configs/production.yml` file is used internally
* Requires `STABLE_TAG` environment set
* Deployed via `make stable`

#### Staging `you.better-call.dev`
* Latest docker images `latest`
* `/configs/you.yml` file is used internally
* Deployed via `make latest`

#### Development `localhost`
* `/configs/development.yml` file is used
* You can spawn local instances of databases and message queue or _ssh_ to staging host with port forwarding
* Run services `make {service}` (where service is one of `api` `indexer` `metrics` `compiler`)

#### Sandbox `bbbox`
* `/configs/sandbox.yml` file is used
* Start via `COMPOSE_PROJECT_NAME=bcd-box docker-compose -f docker-compose.sandbox.yml up -d --build`
* Stop via `COMPOSE_PROJECT_NAME=bcd-box docker-compose -f docker-compose.sandbox.yml down`


## Running

### Startup
It takes around 20-30 seconds to initialize all services, API endpoints might return errors until then.  
**NOTE** that if you specified local RPC node that's not running, BCDHub will wait for it indefinitely.

## Snapshots
Full indexing process requires about 2 hours, however there are cases when you cannot afford that.  
Elastic Search has a built-in incremental snapshotting mechanism which we use together with the AWS S3 plugin.

**NOTE:** currently we don't provide public snapshots.  
You can set up your own S3 repo: https://medium.com/@federicopanini/elasticsearch-backup-snapshot-and-restore-on-aws-s3-f1fc32fbca7f  
Alternatively, contact us for granting access

### Get ready
* Make sure you have snapshot settings in your `.env` file
* Elastic service should be up and initialized

### Make snapshot

#### 1. Initialize credentials
```
make s3-creds
```
No further actions required

#### 2. Create local repository (if not exists)
```
make s3-repo
```
Follow the instruction: you can choose an arbitrary name for your repo.

#### 3. Create snapshot
```
make s3-snapshot
```
Select an existing repository to store your snapshot.

#### 4. Schedule automatic snapshots
```
make s3-policy
```
Select an existing repository and configure time intervals using cron expressions: https://www.elastic.co/guide/en/elasticsearch/reference/master/trigger-schedule.html#schedule-cron

### Restore snapshot

#### 1. Cleanup (optional)
In some cases it's not possible to apply snapshot on top of existing indices. You'd need to clear the data then.  
**WARNING:** This will delete all data from your elastic instance.
```
make es-reset
```
Wait for Elastic to be initialized.

#### 2. Initialize creds and repo
Follow steps 1 and 2 from the _make snapshot_ instruction.

#### 3. Apply snapshot
```
make s3-restore
```
Select the latest (by date) snapshot from the list. It's taking a while, don't worry about the seeming freeze.

## Version upgrade
This is mostly for production environment, for all others a simple "start from the scratch" would work.

### Soft update
E.g. applying hotfixes. No breaking changes in the database schema.

#### 1. Build stable images
Make sure you are on master branch
```
git pull
make stable-images
make stable
```

#### 1'. Pull stable images
```
make stable-pull
```

#### 2. Deploy
```
make stable
```

### Data migration
E.g. new field added to one of the elastic models. You'd need to write a migration script to update existing data.

#### 1. Pull migration script
```
git pull
```

#### 2. Run migration
```
make migration
```
Select your script.


### Upgrade from snapshot
In case you need to reindex from scratch you can set up a secondary BCDHub instance, fill the index, make a snapshot, and then apply it to the production instance.

#### 0. Make a snapshot
Typically you'd use staging for that.

#### 1. Stop BCDHub and clear indexed data
```
make upgrade
```
Wait for Elastic to be initialized after restart.

#### 2. Restore snapshot
```
make s3-restore
```
Select the snapshot you made.

#### 3. Run the rest of the services
```
make stable
```

## Contact us
* Telegram: https://t.me/baking_bad_chat
* Twitter: https://twitter.com/TezosBakingBad
* Slack: https://tezos-dev.slack.com/archives/CV5NX7F2L


## About
This project is the successor of the first [serverless](https://github.com/baking-bad/better-call-dev) version (aka BCD1). It has been rewritten from scratch in Golang.   
Better Call Dev was initially funded and is currently supported by [Tezos Foundation](https://tezos.foundation/).
