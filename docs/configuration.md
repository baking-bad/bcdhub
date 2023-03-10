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
        requests_per_second: 10
```

#### `tzkt`
TzKT API endpoints (optional) and connection timeouts
```yml
tzkt:
    mainnet:
        uri: https://api.tzkt.io/v1/
        base_uri: https://api.tzkt.io
        timeout: 20
```

#### `db`
PostgreSQL connection string
```yml
storage:
  pg: 
    host: ${DB_HOSTNAME:-db}
    port: 5432
    user: ${POSTGRES_USER}
    dbname: ${POSTGRES_DB:-bcd}
    password: ${POSTGRES_PASSWORD}
    sslmode: disable
  timeout: 10
```

#### `sentry`
[Sentry](https://sentry.io/) configuration
```yml
sentry:
  environment: production
  uri: ${SENTRY_DSN}
  front_uri: ${SENTRY_DSN_FRONT}
  debug: false
```

#### `share_path`
Folder to store cached contract sources
```yml
share_path: /etc/bcd
```

#### `services`
Some third-party services
```yml
  services:
    mainnet:
        mempool: https://mempool.dipdup.net/v1/graphql
```

#### `api`
API service settings
```yml

api:
  project_name: api
  bind: ":14000"
  cors_enabled: false
  sentry_enabled: true
  seed_enabled: false
  page_size: ${PAGE_SIZE:-10}
  frontend:
    ga_enabled: true
    mempool_enabled: true
    sandbox_mode: false
    rpc:
      mainnet: https://rpc.tzkt.io/mainnet
  networks:
    - mainnet
  connections:
    max: 50
    idle: 10
```

#### `indexer`
Indexer service settings.
```yml
indexer:
  project_name: indexer
  sentry_enabled: true
  networks:
    mainnet:
      receiver_threads: ${MAINNET_THREADS:-1}
  connections:
    max: 5
    idle: 5
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
  connections:
    max: 5
    idle: 5

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

#### Credentials _required_
* `POSTGRES_USER` e.g. _root_
* `POSTGRES_PASSWORD` e.g. _root_
* `POSTGRES_DB` e.g. _bcd_

#### Services ports _required_
* `BCD_API_PORT` e.g. _14000_
* `POSTGRES_PORT` e.g. _5432_
* `BCD_GUI_PORT` e.g. _8000_

#### Sentry creds _required if `sentry_enabled: true`_
* `SENTRY_DSN`

#### AWS settings
* `AWS_ACCESS_KEY_ID`
* `AWS_SECRET_ACCESS_KEY`

#### Others
* `STABLE_TAG` _required for building & running images_ e.g. _2.5_