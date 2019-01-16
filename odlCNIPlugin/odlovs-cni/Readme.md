# odlovs-cni

## Building the binary

We assume you already installed golang and dep; if not, check the below links for more info:

1. install golang : https://golang.org/doc/install
1. install dep : https://github.com/golang/dep

Run `make binary` to download the dependencies and build odlovs-cni.
The `odlovs-cni` binary will be under the `bin/` directory.


## Building the container

Run `make container` to build the container image.

In order to let odlovs-cni container image run properly in a K8s cluster, the following things should be considered:

1. The k8s cluster config file should be exist under $HOME/.kube/config in each cluster node.
1. The Pod-cidr should be set in the K8s cluster configuration.

Use the `container/odlcni.yaml` like this:
   - $ kubectl create -f odlcni.yaml
