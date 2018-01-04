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
	"net"
	"runtime"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/j-keck/arping"
	"github.com/vishvananda/netlink"
	"os"
	"strings"
	"time"
)

const (
	mtu     = 1400
	netmask = "/24"
)

func cmdAdd(args *skel.CmdArgs) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ovsConfig, err := parseOdlCniConf(args.StdinData)
	if err != nil {
		return fmt.Errorf("Error while parse conf: %v", err)
	}
	// Get Open vSwitch driver
	ovsDriver := NewOvsDriver(ovsConfig.OvsBridge)
	// sleep to make sure the bridge link has been created
	time.Sleep(300 * time.Millisecond)

	if ovsConfig.CtlrActive {
		ovsDriver.SetActiveController(ovsConfig.Controller.String(), ovsConfig.CtlrPort)
	} else {
		ovsDriver.SetPassiveController(ovsConfig.CtlrPort)
	}
	if ovsConfig.MgrActive {
		ovsDriver.SetActiveManager(ovsConfig.Manager.String(), ovsConfig.MgrPort)
	  } else {
		ovsDriver.SetPassiveManager(ovsConfig.MgrPort)
	}

	// Get Container network namespace
	contNetNS, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("Error open netns %q: %v", args.Netns, err)
	}
	defer contNetNS.Close()

	// Setup the contNetNS tap and set container interface
	contIface := &current.Interface{}
	hostIface := &current.Interface{}

	err = contNetNS.Do(func(hostNS ns.NetNS) error {
		// create the veth pair in the container and move host end into host netns
		hostVeth, containerVeth, err := ip.SetupVeth(args.IfName, mtu, hostNS)
		if err != nil {
			return fmt.Errorf("Error Setup Veth, %v", err)
		}
		contIface.Name = containerVeth.Name
		contIface.Mac = containerVeth.HardwareAddr.String()
		contIface.Sandbox = contNetNS.Path()
		hostIface.Name = hostVeth.Name
		hostIface.Mac = hostVeth.HardwareAddr.String()
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error while setup the container NetNS, %v", err)
	}

	k8sArgs := K8sArgs{}
	err = types.LoadArgs(args.Args, &k8sArgs)
	if err != nil {
		return fmt.Errorf("Error while parsing k8s arguments, ", err)
	}

	err = ovsDriver.CreatePort(hostIface.Name, "", 0, string(k8sArgs.K8S_POD_NAMESPACE+":"+k8sArgs.K8S_POD_NAME))
	if err != nil {
		return fmt.Errorf("Error adding created pods veth to ovs bridge %v", err)
	}

	// We use the default CNI IPAM for now till we decide how will use ODL dhcp service
	// Run the IPAM plugin and get back the config to apply
	r, err := ipam.ExecAdd(ovsConfig.IPAM.Type, args.StdinData)
	if err != nil {
		return fmt.Errorf("Error execAdd IPAM plugin, %v", err)
	}
	// Convert whatever the IPAM result into the current Result type
	result, err := current.NewResultFromResult(r)
	if err != nil {
		return fmt.Errorf("Error convert the IPAM result into current Result, %v", err)
	}
	if len(result.IPs) == 0 {
		return fmt.Errorf("Error IPAM plugin returned missing IP config")
	}

	result.Interfaces = []*current.Interface{hostIface, contIface}

	// Configure the container hardware and IP addresses
	if err := contNetNS.Do(func(_ ns.NetNS) error {
		contIface, err := net.InterfaceByName(args.IfName)
		if err != nil {
			return fmt.Errorf("Error getting the conatiner ifName, %v", err)
		}

		// Add the IP to the interface
		if err := ConfigureIface(contIface.Name, result); err != nil {
			return fmt.Errorf("Error Adding IpAddress to ifName, %v", err)
		}

		// Just for now send arp to all other ports. Will delete this once ctlr push
		// flow rules to the bridge.
		for _, ipc := range result.IPs {
			if ipc.Version == "4" {
				_ = arping.GratuitousArpOverIface(ipc.Address.IP, *contIface)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("Error configure container Hardware And IP Addresses, %v", err)
	}

	// Add the public interface to ovs bridge
	if ovsConfig.ExternalIntf != "" {
		err := ovsDriver.CreatePort(ovsConfig.ExternalIntf, "", 0, "")
		if err != nil {
			return fmt.Errorf("Error Adding external net interface %v", err)
		}
	}
	// Set the default gw to the ovsbrk8s intf
	link, _ := netlink.LinkByName(ovsConfig.OvsBridge)
	if link.Attrs().OperState != netlink.OperUp {
		cidr := result.IPs[0].Gateway.String()
		if strings.IndexByte(cidr, '/') < 0 {
			cidr = cidr + netmask
		}
		ipNet, err := netlink.ParseIPNet(cidr)
		if err != nil {
			return fmt.Errorf("Error parsing external IPAddress %v", err)
		}
		addr := &netlink.Addr{
			IPNet: ipNet,
			Label: "",
			Flags: 0,
			Scope: 0,
		}
		netlink.AddrAdd(link, addr)
		netlink.LinkSetUp(link)
	}

	return types.PrintResult(result, ovsConfig.CNIVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	ovsConfig, err := parseOdlCniConf(args.StdinData)
	if err != nil {
		return fmt.Errorf("Error while parse conf: %v", err)
	}
	if err := ipam.ExecDel(ovsConfig.IPAM.Type, args.StdinData); err != nil {
		return err
	}
	// Get Open vSwitch driver
	ovsDriver := NewOvsDriver(ovsConfig.OvsBridge)
	k8sArgs := K8sArgs{}
	err = types.LoadArgs(args.Args, &k8sArgs)
	if err != nil {
		return fmt.Errorf("Error while parsing k8s arguments, ", err)
	}
	prtName := ovsDriver.GetPortNameByExternalId(string(k8sArgs.K8S_POD_NAMESPACE + ":" + k8sArgs.K8S_POD_NAME))
	return ovsDriver.DeletePortByName(prtName)
}

func main() {
	skel.PluginMain(cmdAdd, cmdDel, version.All)
}

// ConfigureIface takes the result of IPAM plugin and
// applies to the ifName interface
func ConfigureIface(ifName string, res *current.Result) error {
	if len(res.Interfaces) == 0 {
		return fmt.Errorf("no interfaces to configure")
	}

	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return fmt.Errorf("failed to lookup %q: %v", ifName, err)
	}

	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("failed to set %q UP: %v", ifName, err)
	}

	var v4gw, v6gw net.IP
	for _, ipc := range res.IPs {
		if ipc.Interface == nil {
			// set the IPConfig to the container Intf
			ipc.Interface = current.Int(1)
		}
		intIdx := *ipc.Interface
		if intIdx < 0 || intIdx >= len(res.Interfaces) || res.Interfaces[intIdx].Name != ifName {
			return fmt.Errorf("failed to add IP addr %v to %q: invalid interface index", ipc, ifName)
		}

		addr := &netlink.Addr{IPNet: &ipc.Address, Label: ""}
		if err = netlink.AddrAdd(link, addr); err != nil {
			return fmt.Errorf("failed to add IP addr %v", err)
		}

		gwIsV4 := ipc.Gateway.To4() != nil
		if gwIsV4 && v4gw == nil {
			v4gw = ipc.Gateway
		} else if !gwIsV4 && v6gw == nil {
			v6gw = ipc.Gateway
		}
	}

	ip.SettleAddresses(ifName, 10)

	// Add the gateway route
	for _, r := range res.Routes {
		routeIsV4 := r.Dst.IP.To4() != nil
		gw := r.GW
		if gw == nil {
			if routeIsV4 && v4gw != nil {
				gw = v4gw
			} else if !routeIsV4 && v6gw != nil {
				gw = v6gw
			}
		}
		if err = ip.AddRoute(&r.Dst, gw, link); err != nil {
			if !os.IsExist(err) {
				return fmt.Errorf("failed to add route %v", err)
			}
		}
	}
	return nil
}