/*
 * Copyright (c) 2018 Kontron Canada Company and others.  All rights reserved.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v1.0 which accompanies this distribution,
 * and is available at http://www.eclipse.org/legal/epl-v10.html
 */

package ovs_ctrl

import (
	"encoding/binary"
	log "github.com/Sirupsen/logrus"
	"net"
	"sync"
	"github.com/serngawy/libOpenflow/openflow13"
	"github.com/serngawy/libOpenflow/protocol"
	ofctrl "github.com/serngawy/libOpenflow/ofctrl"
	ovs "github.com/serngawy/libovsdb/ovsDriver"
	"git.opendaylight.org/gerrit/p/coe.git/odlKubeProxy/watchers"
	"git.opendaylight.org/gerrit/p/coe.git/odlKubeProxy/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	toCtrlFlowPriority = 10
	toTunnelsFlowPriority = 20
	srvFlowPriority = 100
	podFlowPriority = 50
	srvFlowIdealTimeOut = 15
	srvFlowTbl = 0
	podFlowTbl = 1
)

// OvsController struct stores information needed by the controller
type OvsController struct {
	nodeIP       net.IP
	nodeName     string
	ovsDriver    *ovs.OvsDriver
	Switch       *ofctrl.OFSwitch
	endpnts      map[string]*utils.EndPointInfo
	services     map[string]*utils.ServiceInfo
	nodes        map[string]string
	pods         map[string]string
	lock         sync.Mutex
}

func NewOvsController(nodeName string, nodeIP net.IP, bridge string, ctrlPort int) *OvsController {
	ovsCtrl := new(OvsController)
	ovsCtrl.nodeIP = nodeIP
	ovsCtrl.ovsDriver = ovs.NewOvsDriver(bridge)
	ovsCtrl.ovsDriver.SetActiveController("127.0.0.1", ctrlPort)
	ovsCtrl.endpnts = make(map[string]*utils.EndPointInfo)
	ovsCtrl.services = make(map[string]*utils.ServiceInfo)
	ovsCtrl.nodes = make(map[string]string)
	ovsCtrl.pods = make(map[string]string)
	ovsCtrl.nodeName = nodeName
	return ovsCtrl
}

func (ovsCtrl *OvsController) OnPodUpdate(podUpdate *watchers.PodUpdate) {
	if podUpdate.Pod != nil && podUpdate.Pod.Spec.NodeName == ovsCtrl.nodeName && podUpdate.Pod.Status.PodIP != "" {
		switch podUpdate.Op {
		case utils.ADD:
			fallthrough
		case utils.UPDATE:
			{
				Ids, ofPort, err := ovsCtrl.ovsDriver.GetExternalIdsOFportNo("ip-address", podUpdate.Pod.Status.PodIP)
				if err != nil {
					log.Debug("Pod Update: %v , %v", err, podUpdate.Pod.Status.PodIP)
					return
				}
				macAddress := Ids["attached-mac"]
				if macAddress != nil {
					hwMac, err := net.ParseMAC(macAddress.(string))
					if err != nil {
						log.Debug("Pod Update: %v ", err)
						return
					}
					ovsCtrl.setPodFlowRule(ofPort, hwMac, net.ParseIP(podUpdate.Pod.Status.PodIP))
				}
			}
		case utils.REMOVE:
			{
				//FIXME: Need to delete the pod flow rule.
			}
		}
	} else if podUpdate.Pod.Status.PodIP != "" && podUpdate.Pod.Spec.NodeName != ovsCtrl.nodeName {
		ovsCtrl.lock.Lock()
		defer ovsCtrl.lock.Unlock()
		switch podUpdate.Op {
		case utils.ADD:
			fallthrough
		case utils.UPDATE:
			{
				ovsCtrl.pods[podUpdate.Pod.Status.PodIP] = podUpdate.Pod.Status.HostIP
			}
		case utils.REMOVE:
			{
				delete(ovsCtrl.pods, podUpdate.Pod.Status.PodIP)
			}
		}
	}
}

func (ovsCtrl *OvsController) OnEndpointsUpdate(endpointsUpdate *watchers.EndpointsUpdate) {
	endpnt := utils.CreateEndpointInfo(endpointsUpdate.Endpoints)
	if endpnt != nil {
		ovsCtrl.lock.Lock()
		defer ovsCtrl.lock.Unlock()
		switch endpointsUpdate.Op {
		case utils.ADD:
			fallthrough
		case utils.UPDATE:
			{
				ovsCtrl.endpnts[endpnt.GetEndPntIdentifier()] = endpnt
				ovsCtrl.setupInitialSrvFlow(ovsCtrl.services[endpnt.GetEndPntIdentifier()], endpnt)
			}
		case utils.REMOVE:
			{
				delete(ovsCtrl.endpnts, endpnt.GetEndPntIdentifier())
			}
		}
	}
}

func (ovsCtrl *OvsController) OnServiceUpdate(servicesUpdate *watchers.ServiceUpdate) {
	srv := utils.CreateServiceInfo(servicesUpdate.Service)
	if srv != nil {
		ovsCtrl.lock.Lock()
		defer ovsCtrl.lock.Unlock()
		switch servicesUpdate.Op {
		case utils.ADD:
			fallthrough
		case utils.UPDATE:
			{
				ovsCtrl.services[srv.GetSrvIdentifier()] = srv
				ovsCtrl.setupInitialSrvFlow(srv, ovsCtrl.endpnts[srv.GetSrvIdentifier()])
			}
		case utils.REMOVE:
			{
				delete(ovsCtrl.services, srv.GetSrvIdentifier())
			}
		}
	}
}

func (ovsCtrl *OvsController) OnNodeUpdate(nodeUpdate *watchers.NodeUpdate) {
	ndIP, ndName := "", ""
	for _, address := range nodeUpdate.Node.Status.Addresses {
		if address.Type == v1.NodeInternalIP {
			ndIP = address.Address
		} else if address.Type == v1.NodeHostName {
			ndName = address.Address
		}
	}
	if ovsCtrl.nodeIP.String() == ndIP || ndIP == "" || ndName == "" {
		return
	}
	if ovsCtrl.nodeName == ndName {
		ovsCtrl.nodeIP = net.ParseIP(ndIP)
		return
	}
	ovsCtrl.lock.Lock()
	defer ovsCtrl.lock.Unlock()
	switch nodeUpdate.Op {
	case utils.ADD:
		{
			opts := make(map[string]string)
			opts["key"] = "flow"
			opts["local_ip"] = ovsCtrl.nodeIP.String()
			opts["remote_ip"] = ndIP
			intfName := "tun" + ndIP
			err := ovsCtrl.ovsDriver.CreatePort(ovsCtrl.ovsDriver.OvsBridgeName, intfName, "vxlan", 0, nil, opts)
			if err != nil {
				log.Error("Error creating tunnel %s, %v", ndIP, err)
			}
			ovsCtrl.nodes[ndName] = ndIP
		}
	case utils.REMOVE:
		{
			err := ovsCtrl.ovsDriver.DeletePortByName(ovsCtrl.ovsDriver.OvsBridgeName, "tun" + ndIP)
			if err != nil {
				log.Error("Error deleting tunnel %s, %v", ndIP, err)
			}
			delete(ovsCtrl.nodes, ndName)
		}
	case utils.UPDATE:
		{
			/*if ovsCtrl.nodes[ndName] != ndIP {
				err := ovsCtrl.ovsDriver.DeletePortByName(ovsCtrl.ovsDriver.OvsBridgeName, "tun" + ndIP)
				if err != nil {
					log.Error("Error update tunnel %s, %v", ndIP, err)
				}
				opts := make(map[string]string)
				opts["key"] = "flow"
				opts["local_ip"] = ovsCtrl.nodeIP.String()
				opts["remote_ip"] = ndIP
				intfName := "tun" + ndIP
				err = ovsCtrl.ovsDriver.CreatePort(ovsCtrl.ovsDriver.OvsBridgeName, intfName, "vxlan", 0, nil, opts)
				if err != nil {
					log.Error("Error update tunnel %s, %v", ndIP, err)
				}
				ovsCtrl.nodes[ndName] = ndIP
			}*/
		}
	}
}

func (ovsCtrl *OvsController) PacketRcvd(sw *ofctrl.OFSwitch, packet *openflow13.PacketIn) {

	if packet.Data.Data != nil {
		data, _ := packet.Data.Data.MarshalBinary()
		if len(data) > 0 {
			sourceIP := net.IPv4(data[12], data[13], data[14], data[15])
			srvIP := net.IPv4(data[16], data[17], data[18], data[19])
			tcpSrcPortNo := binary.BigEndian.Uint16(data[20:22])
			tcpDstPortNo := binary.BigEndian.Uint16(data[22:24])
			endpnt, srv := ovsCtrl.findEndPntSrv(srvIP, int32(tcpDstPortNo))
			if endpnt == nil {
				log.Debugln("Packet Rcvd, No endpoint associated %v:%v", srvIP, tcpDstPortNo)
				return
			}
			ofDestPortNo, _ := ovsCtrl.ovsDriver.GetOfPortNoByExternalId("iface-id",
				endpnt.GetPodIdentifier())

			srcHwMac := net.HardwareAddr{}
			ofSrcPortNo, _ := ovsCtrl.ovsDriver.GetOfPortNoByExternalId("ip-address", sourceIP.String())
			if ofSrcPortNo == 0 {
				ndIP := ovsCtrl.pods[sourceIP.String()]
				if ndIP != "" {
					ofSrcPortNo, _ = ovsCtrl.ovsDriver.GetTunnelPortNoByRemoteIP(ndIP)
				} else {
					ofSrcPortNo, _ = ovsCtrl.ovsDriver.GetOfPortNoByExternalId("ip-address", srvIP.String())
					srcHwMac = packet.Data.HWSrc
					if ofSrcPortNo == 0 {
						log.Println("Can not find port number")
						return
					}
				}
			}
			// Ids of dest Pod
			Ids, _ := ovsCtrl.ovsDriver.GetExternalIds("iface-id", endpnt.GetPodIdentifier())
			macAddress := Ids["attached-mac"]
			dstHwMac, _ := net.ParseMAC(macAddress.(string))
			temp := Ids["ip-address"]
			podIP := net.ParseIP(temp.(string))
			ovsCtrl.setEndPointFlowRule(ofDestPortNo, podIP,srvIP, endpnt.Ports[0].Protocol, int32(tcpDstPortNo),
				srv.Ports[0].TargetPort, dstHwMac, sourceIP, ofSrcPortNo, tcpSrcPortNo, srcHwMac)
		}
	}
}

func (ovsCtrl *OvsController) SwitchConnected(sw *ofctrl.OFSwitch) {
	log.Printf("App: Switch connected: %v", sw.DPID())
	ovsCtrl.Switch = sw
	ovsCtrl.initPipeline()
}

func (ovsCtrl *OvsController) SwitchDisconnected(sw *ofctrl.OFSwitch) {
	log.Printf("App: Switch disconnected: %v", sw.DPID())
}

func (ovsCtrl *OvsController) MultipartReply(sw *ofctrl.OFSwitch, rep *openflow13.MultipartReply) {
	log.Debugln(rep.Body)
}


func (ovsCtrl *OvsController) PortStatusChange(sw *ofctrl.OFSwitch, portStatus *openflow13.PortStatus) {
	log.Debugln("Port state: %+v", portStatus)
}

func (ovsCtrl *OvsController) FlowRemoved(sw *ofctrl.OFSwitch, flowRemoved *openflow13.FlowRemoved) {
	log.Debugln("Flow removed: %+v", flowRemoved)
}

func (ovsCtrl *OvsController) PopulateNodes(ndList *v1.NodeList) {
	for _, nd := range ndList.Items {
		wNd := watchers.NodeUpdate {
			Node: &nd,
			Op:utils.ADD,
		}
		ovsCtrl.OnNodeUpdate(&wNd)
	}
}

func (ovsCtrl *OvsController) PopulateResources(k8s_client *kubernetes.Clientset) {
	// Populate the Pipline with current existing resources
	// 1- Nodes
	ndList, _ := k8s_client.CoreV1().Nodes().List(metav1.ListOptions{})
	for _, nd := range ndList.Items {
		wNd := watchers.NodeUpdate {
			Node: &nd,
			Op:utils.ADD,
		}
		ovsCtrl.OnNodeUpdate(&wNd)
	}
	//2- Pods
	podList, _ := k8s_client.CoreV1().Pods("").List(metav1.ListOptions{})
	log.Println("pods ", podList)
	for _, pod := range podList.Items {
		wPod := watchers.PodUpdate {
			Pod: &pod,
			Op:utils.ADD,
		}
		ovsCtrl.OnPodUpdate(&wPod)
	}
	//3- EndPoints
	endPntList, _ := k8s_client.CoreV1().Endpoints("").List(metav1.ListOptions{})
	log.Println("endpoints ", endPntList)
	for _, endPnt := range endPntList.Items {
		wEndpnt := watchers.EndpointsUpdate {
			Endpoints: &endPnt,
			Op:utils.ADD,
		}
		ovsCtrl.OnEndpointsUpdate(&wEndpnt)
	}
	//4- Services
	srvList, _ := k8s_client.CoreV1().Services("").List(metav1.ListOptions{})
	log.Println("srvs ", srvList)
	for _, srv := range srvList.Items {
		wSrv := watchers.ServiceUpdate {
			Service: &srv,
			Op:utils.ADD,
		}
		ovsCtrl.OnServiceUpdate(&wSrv)
	}
}

func (ovsCtrl *OvsController) initPipeline() {
	//resubmit to table 1
	flow := ofctrl.NewFlow(srvFlowTbl)
	flow.SetGotoTableAction(1)
	flow.Match.Priority = 0
	ovsCtrl.Switch.InstallFlow(flow)

	//set normal action on table 1
	flow = ofctrl.NewFlow(podFlowTbl)
	flow.SetNormalAction()
	flow.Match.Priority = 0
	ovsCtrl.Switch.InstallFlow(flow)
}

func (ovsCtrl *OvsController) setupInitialSrvFlow(srv *utils.ServiceInfo, endPnt *utils.EndPointInfo) {
	if srv == nil {
		log.Println("Service cannot be nil ")
		return
	}
	if endPnt == nil {
		log.Println("No endpoint asscoiated to the service ", srv.GetSrvIdentifier())
		return
	}
	portNo, _ := ovsCtrl.ovsDriver.GetOfPortNoByExternalId("iface-id", endPnt.GetPodIdentifier())
	if portNo == 0 {
		portNo, err := ovsCtrl.ovsDriver.GetTunnelPortNoByRemoteIP(ovsCtrl.nodes[*endPnt.NodeName])
		if err != nil {
			log.Error("Intial Srv Flow: Error getting tunnel port ", err)
			return
		}
		ovsCtrl.forwardSrvTrafficTunnels(srv.SrvIP, endPnt.Ports[0].Protocol,
					srv.Ports[0].PortNo, uint32(portNo))
	} else {
		switch srv.SrvType {
			case "ClusterIP":
				ovsCtrl.forwardSrvTrafficCtlr(srv.SrvIP, endPnt.Ports[0].Protocol, srv.Ports[0].PortNo)
				for _, ip := range srv.SrvExternalIP {
					ovsCtrl.forwardSrvTrafficCtlr(ip, endPnt.Ports[0].Protocol, srv.Ports[0].PortNo)
				}
			case "NodePort":
				ovsCtrl.forwardSrvTrafficCtlr(srv.SrvIP, endPnt.Ports[0].Protocol, srv.Ports[0].PortNo)
				ovsCtrl.forwardSrvTrafficCtlr(ovsCtrl.nodeIP, endPnt.Ports[0].Protocol, srv.Ports[0].NodePort)
				for _, ip := range srv.SrvExternalIP {
					ovsCtrl.forwardSrvTrafficCtlr(ip, endPnt.Ports[0].Protocol, srv.Ports[0].PortNo)
				}
			case "LoadBalancer":
				log.Println("UnSupported k8s service type")
			case "ExternalName":
				log.Println("UnSupported k8s service type")
			default:
				log.Println("Unknown k8s service type: ", srv.SrvType)
		}
	}
}

func (ovsCtrl *OvsController) setEndPointFlowRule(ofDestPortNo float64, podIPAddress net.IP, srvIPAddress net.IP,
			protcolType string, srcPort int32, dstPort int32, destMacAddress net.HardwareAddr,
			sourceIP net.IP, ofSrcPortNo float64, tcpSrcPortNo uint16, srcMacAddress net.HardwareAddr) {
	flow := ofctrl.NewFlow(srvFlowTbl)
	flow.Match.Priority = srvFlowPriority
	flow.Match.IpDa = &srvIPAddress
	flow.Match.IpSa = &sourceIP
	flow.Match.Ethertype = protocol.IPv4_MSG
	if protcolType == "TCP" || protcolType == "tcp" {
		flow.Match.IpProto = ofctrl.IP_PROTO_TCP
		flow.Match.TcpDstPort = uint16(srcPort)
		flow.Match.TcpSrcPort = tcpSrcPortNo
	}
	if protcolType == "UDP" || protcolType == "udp" {
		flow.Match.IpProto = ofctrl.IP_PROTO_UDP
		flow.Match.UdpDstPort = uint16(srcPort)
		flow.Match.UdpSrcPort = tcpSrcPortNo
	}

	flow.SetL4Field(uint16(dstPort), "TCPDst")
	flow.SetIPField(podIPAddress, "Dst")
	flow.SetMacDa(destMacAddress)
	flow.SetOutputPortAction(uint32(ofDestPortNo))
	flow.IdleTimeout = srvFlowIdealTimeOut
	ovsCtrl.Switch.InstallFlow(flow)

	// The opposite flow
	flow = ofctrl.NewFlow(srvFlowTbl)
	flow.Match.Priority = srvFlowPriority
	flow.Match.IpSa = &podIPAddress
	flow.Match.IpDa = &sourceIP
	flow.Match.Ethertype = protocol.IPv4_MSG
	if protcolType == "TCP" || protcolType == "tcp" {
		flow.Match.IpProto = ofctrl.IP_PROTO_TCP
		flow.Match.TcpSrcPort = uint16(dstPort)
		flow.Match.TcpDstPort = tcpSrcPortNo
	}
	if protcolType == "UDP" || protcolType == "udp" {
		flow.Match.IpProto = ofctrl.IP_PROTO_UDP
		flow.Match.UdpDstPort = uint16(srcPort)
		flow.Match.UdpSrcPort = tcpSrcPortNo
	}

	flow.SetL4Field(uint16(srcPort), "TCPSrc")
	flow.SetIPField(srvIPAddress, "Src")
	if srcMacAddress.String() != "" {
		flow.SetMacDa(srcMacAddress)
	}
	flow.SetOutputPortAction(uint32(ofSrcPortNo))
	flow.IdleTimeout = srvFlowIdealTimeOut
	ovsCtrl.Switch.InstallFlow(flow)
}

func (ovsCtrl *OvsController) setPodFlowRule(ofPort float64, podMacAddress net.HardwareAddr, podIpAddress net.IP) {
	flow := ofctrl.NewFlow(podFlowTbl)
	flow.Match.Priority = podFlowPriority
	flow.Match.MacDa = &podMacAddress
	flow.Match.Ethertype = protocol.ARP_MSG
	flow.SetOutputPortAction(uint32(ofPort))
	ovsCtrl.Switch.InstallFlow(flow)

	flow = ofctrl.NewFlow(podFlowTbl)
	flow.Match.Priority = podFlowPriority
	flow.Match.IpDa = &podIpAddress
	flow.Match.Ethertype = protocol.IPv4_MSG
	flow.SetMacDa(podMacAddress)
	flow.SetOutputPortAction(uint32(ofPort))
	ovsCtrl.Switch.InstallFlow(flow)
}

func (ovsCtrl *OvsController) forwardSrvTrafficCtlr(srvIPAddress net.IP, protcolType string, srcTcpPort int32) {
	flow := ofctrl.NewFlow(srvFlowTbl)
	flow.Match.Priority = toCtrlFlowPriority
	flow.Match.IpDa = &srvIPAddress
	flow.Match.Ethertype = protocol.IPv4_MSG
	if protcolType == "TCP" || protcolType == "tcp" {
		flow.Match.IpProto = ofctrl.IP_PROTO_TCP
		flow.Match.TcpDstPort = uint16(srcTcpPort)
	}
	if protcolType == "UDP" || protcolType == "udp" {
		flow.Match.IpProto = ofctrl.IP_PROTO_UDP
		flow.Match.UdpDstPort = uint16(srcTcpPort)
	}
	flow.SetGotoControllerAction()
	ovsCtrl.Switch.InstallFlow(flow)
}

func (ovsCtrl *OvsController) forwardSrvTrafficTunnels(srvIPAddress net.IP, protcolType string,
				srcTcpPort int32, tunnelPort uint32) {
	flow := ofctrl.NewFlow(srvFlowTbl)
	flow.Match.Priority = toTunnelsFlowPriority
	flow.Match.IpDa = &srvIPAddress
	flow.Match.Ethertype = protocol.IPv4_MSG
	if protcolType == "TCP" || protcolType == "tcp" {
		flow.Match.IpProto = ofctrl.IP_PROTO_TCP
		flow.Match.TcpDstPort = uint16(srcTcpPort)
	}
	if protcolType == "UDP" || protcolType == "udp" {
		flow.Match.IpProto = ofctrl.IP_PROTO_UDP
		flow.Match.UdpDstPort = uint16(srcTcpPort)
	}

	flow.SetOutputPortAction(tunnelPort)
	ovsCtrl.Switch.InstallFlow(flow)
}

func (ovsCtrl *OvsController) setupTunnelFlowRules(tunnelPort uint32) {
	flow := ofctrl.NewFlow(srvFlowTbl)
	flow.Match.Priority = toTunnelsFlowPriority
	flow.Match.InputPort = tunnelPort
	flow.SetGotoTableAction(podFlowTbl)
	ovsCtrl.Switch.InstallFlow(flow)
}

func (ovsCtrl *OvsController) findEndPntSrv(ip net.IP, portNum int32) (*utils.EndPointInfo, *utils.ServiceInfo) {
	if ovsCtrl.nodeIP.Equal(ip) {
		for _, srv := range ovsCtrl.services {
			for _, port := range srv.Ports {
				if port.NodePort == portNum {
					return ovsCtrl.endpnts[srv.GetSrvIdentifier()], srv
				}
			}
		}
	} else {
		for _, srv := range ovsCtrl.services {
			if srv.SrvIP.Equal(ip) {
				return ovsCtrl.endpnts[srv.GetSrvIdentifier()], srv
			}
			for _, extIP := range srv.SrvExternalIP {
				if extIP.Equal(ip) {
					return ovsCtrl.endpnts[srv.GetSrvIdentifier()], srv
				}
			}
		}
	}
	return nil, nil
}