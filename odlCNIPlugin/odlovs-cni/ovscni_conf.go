/*
 * Copyright (c) 2017 Kontron Canada and others.  All rights reserved.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v1.0 which accompanies this distribution,
 * and is available at http://www.eclipse.org/legal/epl-v10.html
 */

package main

import (
	"encoding/json"
	"fmt"
	"github.com/containernetworking/cni/pkg/types"
	"net"
)

//Example of the expected json
//{
//    "cniVersion":"0.3.0",
//    "name":"sym-cni",
//    "type":"symOvSCNI",
//    "mgrPort":6640,
//    "mgrActive":true,
//    "manager":"192.168.33.1",
//    "ovsBridge":"br-int",
//    "ovsExtBridge":"br-ext",
//    "ctlrPort":6653,
//    "ctlrActive":true,
//    "controller":"192.168.33.1",
//    "externalIntf":"eth2",
//    "externalIp":"192.168.50.11",
//    "ipam":{
//        "type":"host-local",
//        "subnet":"10.11.1.0/24",
//        "routes":[{
//        "dst":"0.0.0.0/0"
//        }],
//        "gateway":"10.11.1.1"
//    }
//}

// The odlcni config type for OVS
type OdlCniConf struct {
	types.NetConf
	MgrPort      int    `json:"mgrPort"`
	MgrActive    bool   `json:"mgrActive"`
	Manager      net.IP `json:"manager"`
	OvsBridge    string `json:"ovsBridge"`
	OvsExtBridge string `json:"ovsExtBridge"`
	CtlrPort     int    `json:"ctlrPort"`
	CtlrActive   bool   `json:"ctlrActive"`
	Controller   net.IP `json:"controller"`
	ExternalIntf string `json:"externalIntf"`
	ExternalIp   net.IP `json:"externalIp"`
}

// K8sArgs is the CNI_ARGS used by Kubernetes
type K8sArgs struct {
	types.CommonArgs
	K8S_POD_NAME      types.UnmarshallableString
	K8S_POD_NAMESPACE types.UnmarshallableString
}

// parse odlcni conf
func parseOdlCniConf(stdin []byte) (OdlCniConf, error) {
	odlCniConf := OdlCniConf{}
	err := json.Unmarshal(stdin, &odlCniConf)
	if err != nil {
		fmt.Errorf("failed to parse odlcni configurations: %v", err)
	}
	return odlCniConf, nil
}
