#!/bin/bash

MASTER_IP=$1
PUBLIC_IP=$2
PUBLIC_SUBNET_MASK=$3
ODL_IP=$4

cat > setup_k8s_master_args.sh <<EOL
MASTER_IP=$1
PUBLIC_IP=$2
PUBLIC_SUBNET_MASK=$3
ODL_IP=$4
EOL

# set the ovs instance in passive mode for now.
sudo ovs-vsctl set-manager tcp:$ODL_IP:6640
# Create a OVS physical bridge and move IP address of enp0s9 to br-enp0s9
echo "Creating physical bridge ..."
sudo ovs-vsctl add-br br-enp0s9
sudo ovs-vsctl add-port br-enp0s9 enp0s9
sudo ip addr flush dev enp0s9
sudo ifconfig br-enp0s9 $PUBLIC_IP netmask $PUBLIC_SUBNET_MASK up
sudo ovs-vsctl set-controller br-enp0s9 tcp:$ODL_IP:6653


# Install CNI
pushd ~/
wget https://github.com/containernetworking/cni/releases/download/v0.5.2/cni-amd64-v0.5.2.tgz
popd
sudo mkdir -p /opt/cni/bin
pushd /opt/cni/bin
sudo tar xvzf ~/cni-amd64-v0.5.2.tgz
popd

# Start k8s daemons
pushd k8s/server/kubernetes/server/bin
echo "Starting kubelet ..."
nohup sudo ./kubelet --api-servers=http://$MASTER_IP:8080 --v=2 --address=0.0.0.0 \
                     --enable-server=true --network-plugin=cni \
                     --cni-conf-dir=/etc/cni/net.d \
                     --cni-bin-dir="/opt/cni/bin/" 2>&1 0<&- &>/dev/null &
sleep 5
popd