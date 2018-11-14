#!/bin/sh

mkdir -p bin
dep ensure
go build -o bin/odlovs-cni

