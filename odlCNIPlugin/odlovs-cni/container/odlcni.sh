#!/bin/sh
cp odlovs-cni /opt/cni/bin/
SERVER=$(hostname)
var=$(kubectl describe node $(echo $SERVER | tr '[:upper:]' '[:lower:]') | grep PodCIDR | awk '{ print $2 }' | cut -d"." -f1-3)
echo "node-name: " $SERVER " Podcidr: " $var .0/24
cat << CNI > /etc/cni/net.d/odlovs-cni.conf
{
    "cniVersion":"0.3.0",
    "name":"odl-cni",
    "type":"odlovs-cni",
    "mgrPort":6640,
    "mgrActive":true,
    "manager":"$mgr_IPAddress",
    "ovsBridge":"br-int",
    "ctlrPort":6653,
    "ctlrActive":true,
    "controller":"$ctrl_IPAddress",
    "externalIntf":"$ext_interface",
    "externalIp":"$ext_IPAddress",
    "ipam":{
        "type":"host-local",
        "subnet":"$var.0/24",
        "routes":[{
            "dst":"0.0.0.0/0"
        }],
        "gateway":"$var.1"
    }
}
CNI