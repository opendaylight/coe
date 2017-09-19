package odl

import (
	"net"

	"k8s.io/apimachinery/pkg/types"
)

type Namespace struct {
	Coe Coe `json:"coe"`
}

type Coe struct {
	Pods    []Pod  `json:"pods"`
	Version string `json:"version,omitempty"`
}

type Pod struct {
	UID        types.UID   `json:"uid"`
	Interfaces []Interface `json:"interface"`
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
	HostName          string    `json:"k8s-node:host-name"`
	InternalIPAddress net.IP    `json:"k8s-node:internal-ip-address,omitempty"`
	ExternalIPAddress net.IP    `json:"k8s-node:external-ip-address,omitempty"`
}
