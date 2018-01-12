==============================
Setting Up COE Dev Environment
==============================

For COE to work end to end, the below modules are required :

#. ODL K8S Watcher
#. OVS CNI plugin
#. ODL netvirt
#. Kubernetes Orchestration

Subsequent sections explain how to set up each of the above modules in a development environment.
A VagrantFile for doing the below steps is already available in coe repository under coe/resources folder.


Building COE Watcher and CNI Plugin
-----------------------------------

Watcher and odlovscni modules reside within ODL COE repository. These modules are written in
golang and for compiling the binaries you need a golang environment.

- install golang

  - download golang from https://golang.org/doc/install?download
  - mkdir $HOME/opt
  - mkdir $HOME/go
  - mkdir $HOME/go/bin
  - tar -xvf golang
  - export GOPATH=$HOME/go
  - export GOROOT=$HOME/opt/go
  - export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

- install glide

  - curl https://glide.sh/get | sh

- install git and clone coe repository

  -  sudo apt install git
  -  go get git.opendaylight.org/gerrit/p/coe.git

- build coe watcher binary

  -  cd $GOPATH/src/git.opendaylight.org/gerrit/coe.git/watcher
  -  glide install
  -  go build
  -  once the above step is completed, a "watcher" binary will be generated in the same folder

- build coe cni plugin

  - cd $GOPATH/src/git.opendaylight.org/gerrit/coe.git/odlCniPlugin/odlovs-cni
  - ./build.sh
  - cni binary "odlovs-cni" will be created under bin directory


Setting Up ODL Netvirt
======================


- clone and build netvirt repository

  - This step is required only if you want to make some changes to the existing code and experiment.
    Else you can just download the latest odl distribution
    go get git.opendaylight.org/gerrit/p/netvirt.git
    cd netvirt
    mvn clean install

- Run ODL
    - cd karaf/target/assembly/bin
    - karaf clean
    - Once the karaf console comes up, install odl-netvirt-coe feature which will bring up all required modules for k8s integration
    - opendaylight-karaf>feature:install odl-netvirt-coe odl-restconf


Setting Up K8S Master and Minions
=================================

VMs need to be setup to run K8S Master and Minions. OVS, Docker and K8S need to be installed on all 3 VMs based on the steps specified below.
The same is available as part of the Vagrant setup script under coe/resources/K8s-ODL-Vagrant/provisioning/setup-Req.sh
[https://github.com/opendaylight/coe/blob/master/resources/K8s-ODL-Vagrant/provisioning/setup-Req.sh]

Install OVS
===========

- sudo apt-get build-dep dkms
- sudo apt-get install -y autoconf automake bzip2 debhelper dh-autoreconf \
                        libssl-dev libtool openssl procps python-six dkms
- git clone https://github.com/openvswitch/ovs.git
- pushd ovs/
- sudo DEB_BUILD_OPTIONS='nocheck parallel=2' fakeroot debian/rules binary
- popd
- sudo dpkg -i openvswitch-datapath-dkms*.deb
- sudo dpkg -i openvswitch-switch*.deb openvswitch-common*.deb \
             python-openvswitch*.deb libopenvswitch*.deb


Install Docker
==============

- sudo apt-get update
- sudo apt-get install -y apt-transport-https ca-certificates
- sudo apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D
- sudo su -c "echo \"deb https://apt.dockerproject.org/repo ubuntu-xenial main\" >> /etc/apt/sources.list.d/docker.list"
- sudo apt-get update
- sudo apt-get purge lxc-docker
- sudo apt-get install -y linux-image-extra-$(uname -r) linux-image-extra-virtual
- sudo apt-get install -y docker-engine bridge-utils
- sudo service docker start

Install Kubernetes
==================

- sudo curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
- sudo su -c "echo \"deb http://apt.kubernetes.io/ kubernetes-xenial main\" >> /etc/apt/sources.list.d/kubernetes.list"
- sudo apt-get update
- sudo apt-get install -y kubelet kubeadm kubernetes-cni

Setting Up K8S CNI Plugin
=========================

The below steps can be found under the ReadMe file at https://github.com/opendaylight/coe/tree/master/resources/K8s-ODL-Vagrant

- sudo mkdir -p /etc/cni/net.d/
- copy the appropriate conf files present in coe repo to net.d folder

  - cd $GOPATH/src/git.opendaylight.org/gerrit/coe.git/resources/example
  - sudo cp master.odlovs-cni.conf /etc/cni/net.d/ [For minions, copy the worker conf file instead of master.conf]

- sudo mkdir -p /opt/cni/bin
- copy the odlovs-cni binary which we compiled from coe repo, to the cni/bin folder.

  - cd $GOPATH/src/git.opendaylight.org/gerrit/coe.git/odlCNIPlugin/odlovs-cni
  - sudo cp odlovs-cni /opt/cni/bin


Start Kubernetes Cluster
========================

- sudo kubeadm init --apiserver-advertise-address={K8S-Master-Node-IP}

  - note: read the command output in order to use the kubectl command after
  - note: in the minion VMs you will use the join command instead ex:
  - vagrant@k8sMinion2:~$ sudo kubeadm join --token {given_token} {K8S-Master-Node-IP}:6443
  - mkdir -p $HOME/.kube
  - sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  - sudo chown $(id -u):$(id -g) $HOME/.kube/config


Start COE Watcher on K8S Master
===============================

- cd $GOPATH/src/git.opendaylight.org/gerrit/coe.git/watcher
- ./watcher odl
- The above step will start the coe watcher, which watches for events from kubernetes, and propagate the same to ODL.

  note : for watcher to start properly, .kube/config file should be setup properly, this will be explained in the output of kubeadm init command.

Bring up PODs and test connectivity
===================================

- You can now bring up pods and see if they are able to communicate to each other. Create two busybox pods like below:

  - kubectl create -f https://github.com/kubernetes/kubernetes/blob/master/hack/testdata/recursive/pod/pod/busybox.yaml
  - Check the status of pods by running kubectl get pods -o wide
  - <faseelak> kubectl get pods -o wide
    <faseelak> NAME       READY     STATUS    RESTARTS   AGE       IP           NODE

    <faseelak> busybox1   1/1       Running   0          20m       10.11.1.33   faseela

    <faseelak> busybox2   1/1       Running   0          1m        10.11.1.34   faseela

- Try pinging from one pod to another

  - kubectl exec -it busybox1 ping 10.11.1.34
