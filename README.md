# BCD Hub
[![Build Status](https://github.com/baking-bad/bcdhub/workflows/build/badge.svg)](https://github.com/baking-bad/bcdhub/actions?query=branch%3Amaster+workflow%3A%22build%22)
[![Docker Build Status](https://img.shields.io/docker/cloud/build/bakingbad/bcdhub-api)](https://hub.docker.com/r/bakingbad/bcdhub-api)
[![made_with golang](https://img.shields.io/badge/made_with-golang-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Backend for the [Better Call Dev](https://better-call.dev) contract explorer & developer dashboard.

## Quickstart

### Run BCD

Clone this repo and cd in:
```
make gateway-images
make gateway
```

API gateway is now available at http://localhost:14000/v1/stats

In order to stop or reset BCD:
```
make gateway-down
make gateway-clear
```

### Sandbox

Make sure your Tezos node is exposed at `0.0.0.0:8732`
```
make sandbox-images
make sandbox
```

Sandbox UI is now available at http://localhost:8000

You can also use a builtin Flextesa instance instead (will be exposed at 8732):
```
make flextesa-sandbox
```

In order to stop or reset sandbox:
```
make sandbox-down
make sandbox-clear
```

## Read more

* [Configuration](./docs/configuration.md)
* [Developer docs](./docs/developer.md)


## Contact us
* Telegram: https://t.me/baking_bad_chat
* Twitter: https://twitter.com/TezosBakingBad
* Slack: https://tezos-dev.slack.com/archives/CV5NX7F2L


## About
This project is the successor of the first [serverless](https://github.com/baking-bad/better-call-dev) version (aka BCD1). It has been rewritten from scratch in Golang.   
Better Call Dev was initially funded and is currently supported by [Tezos Foundation](https://tezos.foundation/).

\c indexer 
\dt