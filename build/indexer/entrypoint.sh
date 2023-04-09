#!/bin/sh

curl https://teztnets.xyz/teztnets.json > /teztnets.json
mondaynet_rpc=$(cat /teztnets.json | jq '. | to_entries | map(select(.key | startswith("monday"))) | map(.value.rpc_url)[0]')
dailynet_rpc=$(cat /teztnets.json | jq '. | to_entries | map(select(.key | startswith("daily"))) | map(.value.rpc_url)[0]')

DAILYNET_RPC=$dailynet_rpc MONDAYNET_RPC=$mondaynet_rpc /go/bin/indexer