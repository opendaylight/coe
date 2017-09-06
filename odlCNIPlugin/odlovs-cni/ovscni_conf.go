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
    "github.com/containernetworking/cni/pkg/types/current"
    "github.com/containernetworking/cni/pkg/version"
)

//Example of the expected json
//`{
//    "name":"odl-cni",
//    "cniVersion":"0.3.0",
//    "plugins":[
//    {
//        "type":"bridge",
//        "bridge":"brk8s",
//        "isGateway":true,
//        "isDefaultGateway":true,
//        "forceAddress":false,
//        "ipMasq":true,
//        "mtu":1400,
//        "hairpinMode":false,
//        "ipam":{
//          "type":"host-local",
//          "subnet":"10.11.0.0/16",
//          "rangeStart":"10.11.1.10",
//          "rangeEnd":"10.11.1.150",
//          "routes":[
//            {
//              "dst":"0.0.0.0/0"
//            }
//          ],
//          "gateway":"10.11.1.1"
//        }
//    },
//    {
//        "type":"odlcoe",
//        "runtimeConfig":{
//            "ovsConfig":{
//                "manager": "{ODL_IP-Address}",
//                "mgrPort": 6640,
//                "mgrActive": true,
//                "ovsBridge": "br-k8s",
//                "controller": "{ODL_IP-Address}",
//                "ctlrPort": 6653,
//                "ctlrActive": true,
//                "vtepIPs":[
//                    "10.11.1.1","10.12.1.1"
//                ]
//            }
//        }
//    }
//    ]
//}`
//`json:""`

type OvsConfig struct {
    CoeType string `json:"type"`
    Manager string `json:"manager"`
    MgrPort int `json:"mgrPort"`
    MgrActive bool `json:"mgrActive"`
    OvsBridge string `json:"ovsBridge"`
    Controller string `json:"controller"`
    CtlrPort int `json:"ctlrPort"`
    CtlrActive bool `json: "ctlrActive"`
    VtepIps []string `json:"vtepIps"`
}

// The odl cni config for OVS instance
type OdlCni struct {
    types.NetConf
    RuntimeConfig struct {
                      OvsConfig OvsConfig `json:"ovsConfig"`
                  } `json:"runtimeConfig"`
    // Based on the CNI Spec the runtime MUST also add a prevResult field to the configuration
    // JSON of any plugin after the first one.
    // https://github.com/containernetworking/cni/blob/master/SPEC.md#network-configuration-list-runtime-examples
    RawPrevResult *map[string]interface{} `json:"prevResult"`
    PrevResult    *current.Result         `json:"-"`
}

type OdlCniConfList struct {
        CniVersion string `json:"cniVersion"`
        Name    string     `json:"name"`
        Plugins []*OdlCni `json:"plugins"`
}

// parse odlcni conf
func parseOdlCniConf(stdin []byte) (OdlCni, error) {
    odlCniConf := OdlCni{}
    err := json.Unmarshal(stdin, &odlCniConf)
    if err != nil {
        fmt.Errorf("failed to parse odlcni configurations: %v", err)
    }
    if odlCniConf.RawPrevResult != nil {
        resultBytes, err := json.Marshal(odlCniConf.RawPrevResult)
        if err != nil {
            return odlCniConf, fmt.Errorf("could not serialize prevResult: %v", err)
        }
        res, err := version.NewResult(odlCniConf.CNIVersion, resultBytes)
        if err != nil {
            return odlCniConf, fmt.Errorf("could not parse prevResult: %v", err)
        }
        odlCniConf.RawPrevResult = nil
        odlCniConf.PrevResult, err = current.NewResultFromResult(res)
        if err != nil {
            return odlCniConf, fmt.Errorf("could not convert result to current version: %v", err)
        }
    }

     if odlCniConf.RuntimeConfig.OvsConfig.OvsBridge == "" {
         odlCniConf.RuntimeConfig.OvsConfig.OvsBridge = DefaultBridgeName
    }
    if odlCniConf.RuntimeConfig.OvsConfig.CtlrPort == 0 {
        odlCniConf.RuntimeConfig.OvsConfig.CtlrPort = DefaultControllerPort
    }
    if odlCniConf.RuntimeConfig.OvsConfig.MgrPort == 0 {
        odlCniConf.RuntimeConfig.OvsConfig.MgrPort = DefaultManagerPort
    }
    return odlCniConf, err
}

// parse netconf list
func parseNetConfList(stdin []byte) (OvsConfig, error) {
    odlCniConfList := OdlCniConfList{}
    err := json.Unmarshal(stdin, &odlCniConfList)
    if err != nil {
        fmt.Errorf("failed to parse NetConf list configurations: %v", err)
    }
    // The config json has odlcni at index 1
    ovsConf := odlCniConfList.Plugins[1].RuntimeConfig.OvsConfig
    if ovsConf.OvsBridge == "" {
        ovsConf.OvsBridge = DefaultBridgeName
    }
    if ovsConf.CtlrPort == 0 {
        ovsConf.CtlrPort = DefaultControllerPort
    }
    if ovsConf.MgrPort == 0 {
        ovsConf.MgrPort = DefaultManagerPort
    }
    return ovsConf, err
}