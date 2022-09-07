# BCD Hub
[![Build Status](https://github.com/baking-bad/bcdhub/workflows/build/badge.svg)](https://github.com/baking-bad/bcdhub/actions?query=branch%3Amaster+workflow%3A%22build%22)
[![made_with golang](https://img.shields.io/badge/made_with-golang-blue.svg)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Backend for the [Better Call Dev](https://better-call.dev) contract explorer & developer dashboard.

## Quickstart

### Sandbox

The simplest way is just to copy the `docker-compose.flexesa.yml` to your project.

Make sure you have the latest images and run the compose:
```
make sandbox-pull
make flextesa-sandbox
```
Sandbox UI is now available at http://localhost:8000


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
* Discord: https://discord.gg/RcPGSdcVSx


## About
This project is the successor of the first [serverless](https://github.com/baking-bad/better-call-dev) version (aka BCD1). It has been rewritten from scratch in Golang.   
Better Call Dev was initially funded and is currently supported by [Tezos Foundation](https://tezos.foundation/).
