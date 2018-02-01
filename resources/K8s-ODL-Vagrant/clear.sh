#!/usr/bin/env bash

br=$1

# clean the env after running $ sudo kubeadem reset
sudo systemctl stop kubelet
sudo rm -rf /var/lib/cni/
sudo rm -rf /var/lib/kubelet/*
sudo rm -rf /etc/cni/net.d/*
sudo ovs-vsctl del-br $br
sudo ovs-vsctl del-manager
