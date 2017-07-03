#!/bin/bash

# install docker
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates
sudo apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D
sudo su -c "echo \"deb https://apt.dockerproject.org/repo ubuntu-xenial main\" >> /etc/apt/sources.list.d/docker.list"
sudo apt-get update
sudo apt-get purge lxc-docker
sudo apt-get install -y linux-image-extra-$(uname -r) linux-image-extra-virtual
sudo apt-get install -y docker-engine
sudo service docker start

# Install OVS and dependencies
sudo apt-get build-dep dkms
sudo apt-get install -y autoconf automake bzip2 debhelper dh-autoreconf \
                        libssl-dev libtool openssl procps \
                        python-six dkms

git clone https://github.com/openvswitch/ovs.git
pushd ovs/
sudo DEB_BUILD_OPTIONS='nocheck parallel=2' fakeroot debian/rules binary
popd
sudo dpkg -i openvswitch-datapath-dkms*.deb
sudo dpkg -i openvswitch-switch*.deb openvswitch-common*.deb \
             python-openvswitch*.deb libopenvswitch*.deb

sudo rm -rf *.deb