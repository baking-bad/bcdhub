# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

BCDHub is the Golang backend for [Better Call Dev](https://better-call.dev), a Tezos smart contract explorer & developer dashboard. It is made of two microservices that share a PostgreSQL database:

* **`indexer`** (`cmd/indexer`) — pulls blocks/operations from a Tezos RPC node, decodes contract calls, storage, big-maps, tickets, etc., and writes them to Postgres. Tracks protocol upgrades and can index multiple networks concurrently.
* **`api`** (`cmd/api`) — a Gin-based JSON REST API that serves indexed data, decoding Michelson on the fly, plus account/contract/bigmap/entrypoint endpoints.

Optional third-party dependency: [TzKT](https://github.com/baking-bad/tzkt) for block lists / mempool / metadata (only relevant for public networks, not sandbox).

## Common commands

```bash
# Run services locally (spins up `db` via docker-compose first)
make api        # cd cmd/api && go run -tags=jsoniter .
make indexer    # cd cmd/indexer && go run .

# Tests
go test ./...                       # unit tests
make test                           # same as above
go test ./internal/bcd/ast/...      # single package
go test ./internal/bcd/ast/ -run TestName   # single test

# Postgres-backed integration tests (internal/postgres/tests) spin up a real
# postgres via testcontainers (github.com/dipdup-io/go-lib/testhelpers) —
# Docker must be running locally for these to pass.

# Lint (must be clean before merging, run in CI)
make lint            # golangci-lint run
golangci-lint run ./internal/bcd/...   # scoped

# Regenerate repository mocks (mockgen, invoked via go:generate directives
# in internal/models/*/repository.go)
go generate ./...

# Sandbox environment (flextesa)
make sandbox-pull
make flextesa-sandbox   # http://localhost:8000
make sandbox-down
make sandbox-clear
```

Service selection at runtime is controlled by the `BCD_ENV` env var (`development`, `production`, `sandbox`, `testnets`), which picks a config file under `configs/` (see `internal/config/config.go:LoadDefaultConfig`). For local dev this is `configs/development.yml`, loaded relative to `cmd/{api,indexer}` (`../../configs/development.yml`).

## Architecture

### Context / dependency injection (`internal/config`)

Both services build a `config.Context` (per-network) via `config.NewContext(network, ...ContextOption)`. Options in `internal/config/options.go` (`WithStorage`, `WithRPC`, `WithMempool`, `WithConfigCopy`, ...) wire up concrete Postgres-backed repositories, the RPC client, etc. onto the context. The API additionally builds a `config.Contexts` map keyed by network (one context per configured network) in `cmd/api/main.go`; handlers pull the right one out of the Gin context via a middleware (`c.MustGet("context").(*config.Context)`).

When adding a new piece of infrastructure (a new repository, a new external client), wire it through a `ContextOption` here rather than constructing it ad hoc in handlers/indexer code.

### Storage layer (`internal/models` + `internal/postgres`)

Storage access follows a repository-interface pattern:
* `internal/models/<entity>/` defines the domain model (bun ORM struct, e.g. `Account` in `internal/models/account/model.go`) and a `Repository` interface (`internal/models/account/repository.go`). Each `repository.go` has a `//go:generate mockgen ...` directive producing a mock in `internal/models/mock/<entity>`.
* `internal/postgres/<entity>/storage.go` implements that interface using [bun](https://bun.uptrace.dev/) against Postgres.
* `internal/postgres/core` wraps `*bun.DB` (connection setup, transactions, page size, query logging).

This split lets business logic (parsers, handlers) depend only on the `models` interfaces, keeping Postgres out of their import graph.

Integration tests for the Postgres implementations live in `internal/postgres/tests` as a single `testify/suite.Suite` (`StorageTestSuite`) that boots a real `timescale/timescaledb` container per run and loads fixtures via `go-testfixtures`.

### Michelson / Tezos primitives (`internal/bcd`)

This is the core domain logic for understanding Tezos contracts:
* `internal/bcd/ast` — typed AST for Michelson types & values (`TypedAst`, `Node` implementations per Michelson primitive: `map.go`, `or.go`, `big_map.go`, `lambda.go`, etc.), plus JSON-schema generation, Miguel-format conversion, and docstring generation.
* `internal/bcd/forge` — binary encoding/decoding ("forging") of Michelson data, matching Tezos' wire format.
* `internal/bcd/base` — low-level untyped node representation shared by forge/ast.
* `internal/bcd/formatter` — pretty-printing Michelson.
* `internal/bcd/tezerrors` — decoding/classifying RPC error payloads.
* `internal/bcd/translator` — converts between representations (e.g. Michelson <-> Michelson-JSON).

### Parsers (`internal/parsers`)

Turns raw RPC responses into stored models:
* `internal/parsers/operations` — decodes operation groups (transactions, originations, etc.) into `models/operation`, `models/bigmapdiff`, `models/ticket` records.
* `internal/parsers/storage` — decodes contract storage per protocol era (`alpha`, `babylon`, ...).
* `internal/parsers/protocols` — protocol-upgrade bookkeeping.
* `internal/parsers/migrations` — contract migrations triggered by protocol upgrades.
* `internal/parsers/contract` — contract origination parsing.
* `internal/parsers/stacktrace` — reconstructs execution stack traces for failed operations (ghost contracts, errors).

### Indexer (`cmd/indexer/indexer`)

`BlockchainIndexer` (one per network) drives a loop: fetch next block via `Receiver` (RPC polling with worker threads), diff against current chain state, detect protocol changes (via `internal/parsers/protocols`), run parsers over operations, and commit everything in a single Postgres transaction (`internal/postgres/core.Transaction`) per block, including stats updates. Rollback handling lives in `internal/rollback`.

### API (`cmd/api`)

Gin router set up in `cmd/api/main.go` (`app.makeRouter`); route handlers live in `cmd/api/handlers/*.go`, one file per resource (account, contract, bigmap, entrypoints, operations, run_code, stats, ...). Request validation structs/tags are in `cmd/api/validations`. Handlers are thin: bind request → call into `config.Context` repositories/services → shape response DTOs (`cmd/api/handlers/responses.go`).

### Node RPC client (`internal/noderpc`)

Wraps the Tezos node RPC (`/chains/main/...`) with rate limiting, timeouts, and optional request logging/caching; `WithWaitRPC` variant blocks/retries until the node is reachable (used at indexer startup).

## Notes

* Module path: `github.com/baking-bad/bcdhub`, Go 1.26.
* `api` builds with `-tags=jsoniter` to swap in `json-iterator/go` for faster JSON.
* Versioning is `X.Y.Z`: `Y` bumps signal a breaking change requiring reindex/resync with the frontend (see `docs/developer.md` for the full release/versioning/snapshot workflow).
* Full config reference (per-service YAML keys, required `.env` vars, docker networking for a local RPC node): `docs/configuration.md`.
