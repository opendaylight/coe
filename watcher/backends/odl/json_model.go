package odl

import (
	"net"

	"k8s.io/apimachinery/pkg/types"
)

type Namespace struct {
	Coe Coe `json:"coe"`
}

type Coe struct {
	Pods []Pod `json:"pods"`
}

type Pod struct {
	UID            types.UID   `json:"uid"`
	Name           string      `json:"name"`
	HostIPAddress  string      `json:"host-ip-address,omitempty"`
	NetworkNS      string      `json:"network-NS"`
	PortMacAddress string      `json:"port-mac-address"`
	Interfaces     []Interface `json:"interface"`
}

type Interface struct {
	UID            types.UID `json:"uid"`
	IPAddress      net.IP    `json:"ip-address,omitempty"`
	NetworkID      string    `json:"network-id"`
	NetworkType    string    `json:"network-type"`
	SegmentationID uint32    `json:"segmentation-id"`
}

type Node struct {
	UID               types.UID `json:"k8s-node:uid"`
	PodCIDR           string    `json:"k8s-node:pod-cidr,omitempty"`
	HostName          string    `json:"k8s-node:host-name,omitempty"`
	InternalIPAddress net.IP    `json:"k8s-node:internal-ip-address,omitempty"`
	ExternalIPAddress net.IP    `json:"k8s-node:external-ip-address,omitempty"`
}

type Service struct {
	UID                   types.UID      `json:"service:uid"`
	Name                  string         `json:"service:name"`
	ClusterIPAddress      net.IP         `json:"service:cluster-ip-address"`
	NetworkNS             string         `json:"service:network-NS"`
	ExternalIPAddress     []net.IP       `json:"service:external-ip-address,omitempty"`
	LoadBalancerIPAddress net.IP         `json:"service:load-balancer-IP,omitempty"`
	IngressIPAddress      []net.IP       `json:"service:ingress-ip-address,omitempty"`
	ServicePorts          []ServicePorts `json:"service:service-ports"`
}

type ServicePorts struct {
	Name       string `json:"service:name"`
	Port       int32  `json:"service:port"`
	TargetPort string `json:"service:target-port"`
	NodePort   int32  `json:"service:node-port"`
}

type EndPoints struct {
	UID               types.UID            `json:"service:uid"`
	Name              string               `json:"service:name"`
	NetworkNS         string               `json:"service:network-NS"`
	EndPointAddresses []EndPointsAddresses `json:"service:endpoint-addresses,omitempty"`
	EndPointPorts     []EndPointsPorts     `json:"service:endpoint-ports,omitempty"`
}

type EndPointsAddresses struct {
	IPAddress net.IP  `json:"service:ip-address"`
	HostName  string  `json:"service:host-name"`
	NodeName  *string `json:"service:node-name"`
}

type EndPointsPorts struct {
	Name string `json:"service:name"`
	Port int32  `json:"service:port"`
}
