#!/usr/bin/env bash

# Example of adding external intf and bridge, the IP-address required to changes based on the host IP
sudo ovs-vsctl add-br br-ext
sudo ovs-vsctl add-port br-ext eth2
sudo ip addr flush dev eth2
sudo ip addr add 192.168.40.12/24 dev br-ext
sudo ip link set br-ext up
sudo ovs-vsctl add-port br-ext prt-int -- set interface prt-int type=patch options:peer=prt-ext
sudo ovs-vsctl add-port br-int prt-ext -- set interface prt-ext type=patch options:peer=prt-int external_ids:ip-address=192.168.40.12
sudo ovs-ofctl add-flow br-ext "table=0, priority=10, tcp, nw_dst=192.168.40.0/24 actions=output:prt-int"
sudo ovs-ofctl add-flow br-ext "table=0, priority=10, udp, nw_dst=192.168.40.0/24 actions=output:prt-int"