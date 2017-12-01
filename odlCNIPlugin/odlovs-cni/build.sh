#!/usr/bin/env bash

mkdir -p bin
mkdir -p vendor/src
export GOPATH=$PWD/vendor
echo $GOPATH
glide update

# glide is wired, delete the src directory
mkdir vendor/src

# for some reasons glide didn't put dependencies under the src folder
for dir in vendor/*; do
  if [ "$dir" != "vendor/src" ]; then
     cp -r $dir vendor/src;
  fi;
  done

# duplicating the library confuse go at the build execution.
rm -rf vendor/src/github.com/containernetworking/plugins/vendor/github.com/containernetworking
rm -rf vendor/src/github.com/containernetworking/plugins/vendor/github.com/vishvananda

go build -o bin/odlovs-cni
