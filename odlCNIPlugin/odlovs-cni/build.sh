#!/bin/sh

mkdir -p bin
dep ensure
GOOS=linux GOARCH=amd64 go build -o bin/odlovs-cni

