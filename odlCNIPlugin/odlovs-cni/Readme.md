# odlovs-cni

## Building the container

Run `make` to build the container image.  This basically just does a (multi-stage) `docker build`.  You need to have a _Docker_ version >= 17.05 for [multi-stage builds](https://docs.docker.com/develop/develop-images/multistage-build/) (else you hit _Error parsing reference: "golang:1 as builder" is not a valid repository/tag: invalid reference format)_).

The built `odlovs-cni` binary is only in the container, not locally available (e.g. nothing in a `bin/` directory).

In order to let odlovs-cni container image run properly in a K8s cluster, the following things should be considered:

1. The k8s cluster config file should be exist under $HOME/.kube/config in each cluster node.
1. The Pod-cidr should be set in the K8s cluster configuration.

Use the `container/odlcni.yaml` like this:

    kubectl create -f odlcni.yaml


## Podman & Buildah instead of Docker

Instead of (a recent >= 17.05 version of Docker), you could also use [podman](https://github.com/containers/libpod) :

    sudo dnf remove "docker*" ; sudo dnf install podman-docker ; make

Or instead of completely removing Docker you can also use them in parallel:

    sudo dnf install podman ; podman build -t  odlovs-cni .

although [as of Jan 2019 there seem to be some issues with Podman](https://github.com/containers/libpod/issues/1973).

Alternatively you can try using [buildah](https://github.com/containers/buildah) :

    sudo dnf install buildah ; buildah bud -t odlovs-cni .`
