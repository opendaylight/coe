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
    //"runtime"
    "github.com/containernetworking/cni/pkg/skel"
    "github.com/containernetworking/cni/pkg/version"
    "github.com/vishvananda/netlink"
    "strings"
)

// Get Linux bridge by name
func getBridgeByName(name string) (*netlink.Bridge, error) {
    link, err := netlink.LinkByName(name)
    if err != nil {
        return link, fmt.Errorf("could not get bridge %q: %v", name, err)
    }
    bridge, ok := link.(*netlink.Bridge)
    if !ok {
        return bridge, fmt.Errorf("link %q already exists but is not a bridge", name)
    }
    return bridge, nil
}

func cmdAdd(args *skel.CmdArgs) error {
    //runtime.LockOSThread()
    //defer runtime.UnlockOSThread()
    ovsConfig, err := parseOdlCniConf(args.StdinData)
    if err != nil {
        return fmt.Errorf("Error while parse conflist: %v", err)
    }
    bridgeName := ovsConfig.RuntimeConfig.OvsConfig.OvsBridge
    fmt.Println("bridgeName is ", bridgeName)
    // Create a Open vSwitch bridge
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
    fmt.Println("linux bridge is ", name)
    br, err := getBridgeByName(name)
    if err != nil {
        return err
    }

    // set link master to ovs bridge
    if err := netlink.LinkSetMaster(ovsbrLink, br); err != nil {
        fmt.Println(err)
        return fmt.Errorf("failed to LinkSetMaster %v", err)
    }

    vtepIPs := ovsConfig.RuntimeConfig.OvsConfig.VtepIps
    if len(vtepIPs) > 0 {
        // Create VxLAN tunnelings
        for i := 0; i < len(vtepIPs); i++ {

            // Create interface name based on IP address in order to make sure it is unique
            intfName := fmt.Sprintf("vtep%s", strings.Replace(vtepIPs[i], ".", "_", -1))

            present, vtapName := ovsDriver.IsVtepPresent(vtepIPs[i])
            if !present || (vtapName != intfName) {
                err := ovsDriver.CreateVtep(intfName, vtepIPs[i])
                if err != nil {
                    return fmt.Errorf("Error creating VTEP port %s. Err: %v", intfName, err)
                }
            }

        }
    }

    return nil
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