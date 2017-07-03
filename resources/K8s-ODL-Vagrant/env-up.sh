#! /bin/bash

if [ ! -d "k8s" ]; then
    mkdir k8s
    cd k8s
    wget https://github.com/kubernetes/kubernetes/releases/download/v1.6.6/kubernetes.tar.gz
    tar xvzf kubernetes.tar.gz
        ./kubernetes/cluster/get-kube-binaries.sh
    mkdir server
    cd server
    tar xvzf ../kubernetes/server/kubernetes-server-linux-amd64.tar.gz
    cd ../../
fi
vagrant up