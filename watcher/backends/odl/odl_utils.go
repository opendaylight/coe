/*
 * Copyright (c) 2017 Kontron Canada Company and others. All rights reserved.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v1.0 which accompanies this distribution,
 * and is available at http://www.eclipse.org/legal/epl-v10.html
 */

package odl

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"k8s.io/client-go/pkg/api/v1"
)

const (
	PodsUrl      = "/restconf/config/pod:coe/pods/"
	NodesUrl     = "/restconf/config/k8s-node:k8s-nodes-info/k8s-nodes/"
	ServicesUrl  = "/restconf/config/service:service-information/services/"
	EndPointsUrl = "/restconf/config/service:endpoints-info/endpoints/"
)

// Setting the Node attributes based on K8s API server doc
// https://kubernetes.io/docs/concepts/architecture/nodes/#addresses
func createNodeStructure(node *v1.Node) []byte {
	odlNodes := make([]Node, 1)
	odlNodes[0] = Node{
		UID:     node.GetUID(),
		PodCIDR: node.Spec.PodCIDR,
	}

	for _, address := range node.Status.Addresses {
		switch address.Type {
		case v1.NodeHostName:
			{
				odlNodes[0].HostName = address.Address
			}
		case v1.NodeInternalIP:
			{
				odlNodes[0].InternalIPAddress = net.ParseIP(address.Address)
			}
		case v1.NodeExternalIP:
			{
				odlNodes[0].ExternalIPAddress = net.ParseIP(address.Address)
			}
		default:
			{
				log.Println("Unknown address type: ", address.Type)
			}
		}
	}

	js, err := json.Marshal(odlNodes)
	if err != nil {
		log.Println("Error while formating k8s node object", err)
	}
	jsStr := `{"k8s-node:k8s-nodes":` + string(js) + "}"
	return []byte(jsStr)
}

func createPodStructure(pod *v1.Pod) []byte {
	interfaces := make([]Interface, 1)
	segmentationIDString := fmt.Sprintf("%s:%s", pod.ClusterName, pod.Namespace)
	segmentationSha256 := sha256.Sum256([]byte(segmentationIDString))
	segmentationHash := make([]byte, 24)
	copy(segmentationHash, segmentationSha256[:])
	segmentationID := binary.LittleEndian.Uint32(segmentationHash)

	interfaces[0] = Interface{
		UID:            pod.GetUID(),
		NetworkID:      "00000000-0000-0000-0000-000000000000",
		NetworkType:    "VXLAN",
		SegmentationID: segmentationID,
		IPAddress:      net.ParseIP(pod.Status.PodIP),
	}
	pods := make([]Pod, 1)
	pods[0] = Pod{
		UID:           pod.GetUID(),
		Name:          pod.GetName(),
		HostIPAddress: pod.Status.HostIP,
		NetworkNS:     pod.Namespace,
		Interfaces:    interfaces,
	}
	coe := Coe{
		Pods: pods,
	}
	js, err := json.Marshal(coe)
	if err != nil {
		log.Println("Error while formating pod object", err)
	}
	return js
}

func createServiceStructure(service *v1.Service) []byte {
	srvPorts := make([]ServicePorts, len(service.Spec.Ports))
	for i := 0; i < len(service.Spec.Ports); i++ {
		srvPorts[i] = ServicePorts{
			Name:     service.Spec.Ports[i].Name,
			NodePort: service.Spec.Ports[i].NodePort,
			Port:     service.Spec.Ports[i].Port,
		}
		srvPorts[i].TargetPort = service.Spec.Ports[i].TargetPort.String()
	}

	exIPs := make([]net.IP, len(service.Spec.ExternalIPs))
	for i := 0; i < len(service.Spec.ExternalIPs); i++ {
		ip := net.ParseIP(service.Spec.ExternalIPs[i])
		if ip != nil {
			exIPs[i] = ip
		}
	}

	ingressIPs := make([]net.IP, len(service.Status.LoadBalancer.Ingress))
	for i := 0; i < len(service.Status.LoadBalancer.Ingress); i++ {
		ip := net.ParseIP(service.Status.LoadBalancer.Ingress[i].IP)
		if ip != nil {
			ingressIPs[i] = ip
		}
	}

	services := make([]Service, 1)
	services[0] = Service{
		UID:                   service.GetUID(),
		Name:                  service.GetName(),
		ClusterIPAddress:      net.ParseIP(service.Spec.ClusterIP),
		ExternalIPAddress:     exIPs,
		IngressIPAddress:      ingressIPs,
		NetworkNS:             service.Namespace,
		LoadBalancerIPAddress: net.ParseIP(service.Spec.LoadBalancerIP),
		ServicePorts:          srvPorts,
	}
	js, err := json.Marshal(services)
	if err != nil {
		log.Println("Error while formating service object", err)
	}
	jsStr := `{"service:services":` + string(js) + "}"
	return []byte(jsStr)
}

func createEndpointStructure(endpoint *v1.Endpoints) []byte {
	endPoints := make([]EndPoints, 1)
	endPoints[0] = EndPoints{
		UID:       endpoint.GetUID(),
		Name:      endpoint.GetName(),
		NetworkNS: endpoint.GetNamespace(),
	}
	if len(endpoint.Subsets) > 0 {
		endPointsAddresses := make([]EndPointsAddresses, len(endpoint.Subsets[0].Addresses))
		for i := 0; i < len(endpoint.Subsets[0].Addresses); i++ {
			endPointsAddresses[i].HostName = endpoint.Subsets[0].Addresses[i].Hostname
			ip := net.ParseIP(endpoint.Subsets[0].Addresses[i].IP)
			if ip != nil {
				endPointsAddresses[i].IPAddress = ip
			}
			endPointsAddresses[i].NodeName = endpoint.Subsets[0].Addresses[i].NodeName
		}
		endPntPorts := make([]EndPointsPorts, len(endpoint.Subsets[0].Ports))
		for i := 0; i < len(endpoint.Subsets[0].Ports); i++ {
			endPntPorts[i].Name = endpoint.Subsets[0].Ports[i].Name
			endPntPorts[i].Port = endpoint.Subsets[0].Ports[i].Port
		}
		endPoints[0].EndPointAddresses = endPointsAddresses
		endPoints[0].EndPointPorts = endPntPorts
	}
	js, err := json.Marshal(endPoints)
	if err != nil {
		log.Println("Error while formating service object", err)
	}
	jsStr := `{"service:endpoints":` + string(js) + "}"
	return []byte(jsStr)
}
