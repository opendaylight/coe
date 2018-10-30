#!/bin/bash
# Install OVS and dependencies
sudo apt-get update
sudo apt-get install -y dh-autoreconf sshpass dkms debhelper autoconf automake libssl-dev pkg-config bzip2 openssl python-all procps python-qt4 python-zopeinterface python-twisted-conch

git clone https://github.com/openvswitch/ovs.git -b branch-2.8
pushd ovs/
sudo DEB_BUILD_OPTIONS='nocheck parallel=8' fakeroot debian/rules binary
popd
sudo dpkg -i openvswitch-datapath-dkms*.deb openvswitch-switch*.deb openvswitch-common*.deb python-openvswitch*.deb libopenvswitch*.deb

sudo rm -rf *.deb

# install docker
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates
sudo apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D
sudo su -c "echo \"deb https://apt.dockerproject.org/repo ubuntu-xenial main\" >> /etc/apt/sources.list.d/docker.list"
sudo apt-get update
sudo apt-get purge lxc-docker
sudo apt-get install -y linux-image-extra-$(uname -r) linux-image-extra-virtual
sudo apt-get install -y docker-engine=17.03.1~ce-0~ubuntu-xenial bridge-utils
sudo service docker start

#install k8s
sudo curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
sudo su -c "echo \"deb http://apt.kubernetes.io/ kubernetes-xenial main\" >> /etc/apt/sources.list.d/kubernetes.list"
sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl kubernetes-cni ebtables socat cri-tools

#create .kube directory
mkdir ~/.kube