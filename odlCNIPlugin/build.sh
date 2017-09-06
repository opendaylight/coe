#!/bin/bash

# FIXME not finish yet

mkdir -p bin

if [ ! -h gopath/src/ ]; then
    mkdir -p gopath/src/
fi

export GOPATH=${PWD}/gopath

go get "github.com/Sirupsen/logrus"
go get "github.com/containernetworking/cni/pkg/skel"
go get "github.com/containernetworking/cni/pkg/version"
go get "github.com/vishvananda/netns"
go get "github.com/vishvananda/netlink"
go get "github.com/socketplane/libovsdb"

PLUGINS="src/ipam_odl/ src/odlcoe/"
for d in $PLUGINS; do
    echo $d
    if [ -d "$d" ]; then
        plugin="$(basename "$d")"
        echo "  $plugin"
        if [ -n "$FASTBUILD" ]
        then
            GOBIN=${PWD}/bin go build -pkgdir $GOPATH/pkg "$@" $d
        else
            go build -o "${PWD}/bin/$plugin" -pkgdir "$GOPATH/pkg" "$@" "$d"
        fi
    fi
done