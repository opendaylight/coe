/*
 * Copyright (c) 2017 Kontron Canada and others.  All rights reserved.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v1.0 which accompanies this distribution,
 * and is available at http://www.eclipse.org/legal/epl-v10.html
 */

package main

import (
    "fmt"
    "time"
    "github.com/containernetworking/cni/pkg/skel"
    "github.com/containernetworking/cni/pkg/version"
    "github.com/containernetworking/cni/pkg/types"
    "github.com/vishvananda/netlink"
    "strings"
)

// Get Linux bridge by name
func getBridgeByName(name string) (*netlink.Bridge, error) {
    link, err := netlink.LinkByName(name)
    if err != nil {
        return nil, fmt.Errorf("could not get bridge %q: %v", name, err)
    }
    bridge, ok := link.(*netlink.Bridge)
    if !ok {
        return bridge, fmt.Errorf("link %q already exists but is not a bridge", name)
    }
    return bridge, nil
}

func cmdAdd(args *skel.CmdArgs) error {
    ovsConfig, err := parseOdlCniConf(args.StdinData)
    if err != nil {
        return fmt.Errorf("Error while parse conflist: %v", err)
    }
    bridgeName := ovsConfig.RuntimeConfig.OvsConfig.OvsBridge

    // Create Open vSwitch bridge
    ovsDriver := NewOvsDriver(bridgeName)
    time.Sleep(300 * time.Millisecond) // sleep to make sure the bridge link has been created
    ovsbrLink, err := netlink.LinkByName(bridgeName)
    if err != nil {
        return fmt.Errorf("could not lookup %q bridge: %v", bridgeName, err)
    }
    // enables the link device
    err = netlink.LinkSetUp(ovsbrLink)
    if  err != nil {
        return fmt.Errorf("Error while enabling ovs bridge link %v", err)
    }

    // Get linux bridge
    name := ovsConfig.PrevResult.Interfaces[0].Name
    br, err := getBridgeByName(name)
    if err != nil {
        return err
    }

    // set link master to ovs bridge
    err = netlink.LinkSetMaster(ovsbrLink, br)
    if err != nil {
        return fmt.Errorf("failed to LinkSetMaster %v", err)
    }

    // We create the initial tunneling between the k8s cluster nodes
    // however, for adding new node to the k8s cluster ODL should ask
    // the odlcni agent to create new vtep with the node IP.
    // We consider a full mesh between the cluster nodes.
    vtepIPs := ovsConfig.RuntimeConfig.OvsConfig.VtepIps
    length := len (vtepIPs)
    for i := 0; i < length; i++ {
        vtepIP := vtepIPs[i].String()
        if vtepIP != "" {
            // Create interface name based on IP address in order to make it readable & unique
            intfName := fmt.Sprintf("vtep%s", strings.Replace(vtepIP, ".", "_", -1))
            present, vtapName := ovsDriver.IsVtepPresent(vtepIP)
            if !present || (vtapName != intfName) {
                err := ovsDriver.CreateVtep(intfName, vtepIP)
                if err != nil {
                    return fmt.Errorf("Error creating VTEP port %s. Err: %v", intfName, err)
                }
            }
        }
    }

    return types.PrintResult(ovsConfig.PrevResult, ovsConfig.CNIVersion)
}

func cmdDel(args *skel.CmdArgs) error {
    _, err := parseOdlCniConf(args.StdinData)
    if err != nil {
        return fmt.Errorf("Error while parse conflist: %v", err)
    }
    return nil
}

func main() {
    skel.PluginMain(cmdAdd, cmdDel, version.All)
}