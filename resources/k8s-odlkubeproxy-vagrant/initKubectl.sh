#!/usr/bin/env bash

# setup the .kube/config for k8s cluster nodes and remove the k8s master node constrain to create pods.
yes | sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
sshpass -p 'vagrant' scp ~/.kube/config vagrant@192.168.30.12:~/.kube/
sshpass -p 'vagrant' scp ~/.kube/config vagrant@192.168.30.13:~/.kube/

#sudo kubectl taint nodes --all node-role.kubernetes.io/master-
#sudo kubectl create -f examples/busybox.yaml
