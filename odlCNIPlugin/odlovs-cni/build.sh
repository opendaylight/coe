#!/usr/bin/env bash

Update=$1

export GOPATH=$PWD/vendor
echo "GOPATH=" $GOPATH

if [ ! -d "vendor/src" ] || [ "$Update" = "update" ]; then
    mkdir -p bin
    mkdir -p vendor/src

    glide update
    # glide is wired, create the src dir and move dependencies under it.
    mkdir vendor/src
    for dir in vendor/*; do
      if [ "$dir" != "vendor/src" ]; then
         cp -r $dir vendor/src;
      fi;
      done

    # duplicating the library confuse go at the build.
    rm -rf vendor/src/github.com/containernetworking/plugins/vendor/github.com/containernetworking
    rm -rf vendor/src/github.com/containernetworking/plugins/vendor/github.com/vishvananda
fi
go build -o bin/odlovs-cni

