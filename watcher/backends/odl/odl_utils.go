/*
 * Copyright (c) 2017 Kontron Canada Company and others.  All rights reserved.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v1.0 which accompanies this distribution,
 * and is available at http://www.eclipse.org/legal/epl-v10.html
 */

package odl

import (
	"encoding/json"
	"fmt"

	"k8s.io/client-go/pkg/api/v1"
)

const (
	PodsUrl      = "/restconf/config/pod:coe/pods/"
	NodesUrl     = "/restconf/config/k8s-node:k8s-nodes-info/k8s-nodes/"
	ServicesUrl  = "/restconf/config/pod:coe/pods/" //FIXME not the right url
	EndPointsUrl = "/restconf/config/pod:coe/pods/" //FIXME not the right url
)

func createNodeStructure(node *v1.Node) []byte {
	odlNodes := make([]Node, 1)
	odlNodes[0] = Node{
		UID:      node.GetUID(),
		HostName: node.GetName(),
	}
	js, err := json.Marshal(odlNodes)
	if err != nil {
		fmt.Println("Error while formating k8s node object", err)
	}
	return js
}

func createPodStructure(pod *v1.Pod) []byte {
	interfaces := make([]Interface, 1)
	interfaces[0] = Interface{
		UID:            pod.GetUID(),
		NetworkID:      "00000000-0000-0000-0000-000000000000",
		NetworkType:    "FLAT",
		SegmentationID: 0,
	}
	pods := make([]Pod, 1)
	pods[0] = Pod{
		UID:        pod.GetUID(),
		Interfaces: interfaces,
	}
	coe := Coe{
		Pods: pods,
	}
	js, _ := json.Marshal(coe)
	return js
}
