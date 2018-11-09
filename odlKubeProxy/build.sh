#!/usr/bin/env bash

Update=$1

echo "GOPATH=" $GOPATH

if [ ! -d "vendor" ] || [ "$Update" = "update" ]; then
    glide update -v
fi

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $GOPATH/bin/odlkubeproxy

