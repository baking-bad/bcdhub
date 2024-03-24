#!/bin/sh

curl https://teztnets.xyz/teztnets.json > /teztnets.json
mondaynet_rpc=$(cat /teztnets.json | jq '. | to_entries | map(select(.key | startswith("monday"))) | map(.value.rpc_url)[0]')

MONDAYNET_RPC=$mondaynet_rpc /go/bin/indexer