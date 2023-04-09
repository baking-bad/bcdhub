#!/bin/sh

ROOT_DIR=/usr/share/nginx/html

if [ -z "$BASE_URL" ]; then
    BASE_URL=http://localhost
fi
if [ -z "$API_URL" ]; then
    API_URL=http://localhost:14000/v1
fi
if [ -z "$IPFS_URL" ]; then
    IPFS_URL=https://ipfs.io
fi

for file in $ROOT_DIR/js/*.js* $ROOT_DIR/index.html;
do
  sed -i 's|__BASE_URL__|'$BASE_URL'|g' $file
  sed -i 's|__VUE_APP_API_URI__|'$API_URL'|g' $file
  sed -i 's|__VUE_APP_IPFS_NODE__|'$IPFS_URL'|g' $file
done

nginx -g 'daemon off;'