/*
 * Copyright (c) 2018 Kontron Canada Company and others.  All rights reserved.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v1.0 which accompanies this distribution,
 * and is available at http://www.eclipse.org/legal/epl-v10.html
 */

package utils

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"io/ioutil"

	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"net"
	"strings"
)

type Operation int

const (
	ADD Operation = iota
	UPDATE
	REMOVE
	SYNCED
)


type PortInfo struct {
	PortName string
	PortNo int32
	Protocol string
	// Extra port info for services
	NodePort int32
	TargetPort int32
}

type EndPointInfo struct {
	NodeName *string
	PodName  string
	PodNs    string
	PodIP    net.IP
	EndPntName string
	EndPntNs   string
	Ports      []PortInfo
}

func (self *EndPointInfo) GetPodIdentifier() string {
	return self.PodNs + ":" + self.PodName
}

func (self *EndPointInfo) GetEndPntIdentifier() string {
	return self.EndPntName + ":" + self.EndPntNs
}

type ServiceInfo struct {
	SrvName string
	SrvNs string
	SrvIP net.IP
	SrvType string
	SrvExternalIP []net.IP
	Ports []PortInfo
}

func (self *ServiceInfo) GetSrvIdentifier() string {
	return self.SrvName + ":" + self.SrvNs
}

// The symkloud cni config for OvS
type kubeConf struct {
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

func ReadKubeConf(path string) kubeConf {
	conf := kubeConf{}
	jsonFile, err := os.Open(path)
	if err != nil {
		log.Println("Error reading the odl cni conf file", err.Error())
		return conf
	}
	defer jsonFile.Close()
	bytes, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(bytes, &conf)
	return conf
}

func CreateEndpointInfo(endPoint *v1.Endpoints) *EndPointInfo {
	endPnt := new(EndPointInfo)
	endPnt.EndPntName = endPoint.Name
	endPnt.EndPntNs = endPoint.Namespace
	if len(endPoint.Subsets) == 0 {
		return nil
	}
	for _,subSet := range endPoint.Subsets {
		for _, address := range subSet.Addresses {
			endPnt.PodIP = net.ParseIP(address.IP)
			endPnt.NodeName = address.NodeName
			if address.TargetRef != nil {
				endPnt.PodName = address.TargetRef.Name
				endPnt.PodNs = address.TargetRef.Namespace
			}
		}
		var ports []PortInfo
		for _, port := range subSet.Ports {
			portInfo := PortInfo{
				PortName: port.Name,
				PortNo: port.Port,
				Protocol: string(port.Protocol),
			}
			ports = append(ports, portInfo)
		}
		endPnt.Ports = ports
	}
	return endPnt
}

func CreateServiceInfo(service *v1.Service) *ServiceInfo {
	srv := new(ServiceInfo)
	srv.SrvName = service.Name
	srv.SrvNs = service.Namespace
	srv.SrvIP = net.ParseIP(service.Spec.ClusterIP)
	srv.SrvType = string(service.Spec.Type)
	var srvPorts []PortInfo
	for _, port := range service.Spec.Ports {
		srvPort := PortInfo {
			PortName: port.Name,
			PortNo: port.Port,
			Protocol: string(port.Protocol),
			NodePort: port.NodePort,
			TargetPort: port.TargetPort.IntVal,
		}
		if srvPort.PortNo != 0 && srvPort.TargetPort != 0 {
			srvPorts = append(srvPorts, srvPort)
		}
	}
	srv.Ports = srvPorts
	var extIPs []net.IP
	for _, extIP := range service.Spec.ExternalIPs {
		ip := net.ParseIP(extIP)
		if ip != nil {
			extIPs = append(extIPs, ip)
		}
	}
	srv.SrvExternalIP = extIPs
	//log.Println("service is ", srv)
	return srv
}


func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func GetClientSetlocal() *kubernetes.Clientset {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "./config", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Println("Error at BuildConfigFromFlags %v", err.Error())
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println("Error at NewForConfig %v", err.Error())
		panic(err.Error())
	}
	return clientset
}

func GetClientSetRemote() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}


func GetHostName() (string, error) {
	hostName, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("Error while getting hostName ", err)
	}
	return strings.ToLower(hostName), nil
}

func GetHostNodeIP(k8s_client *kubernetes.Clientset, hostName string) (string, *v1.NodeList, error) {
	ndList, err := k8s_client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return "", ndList, err
	}
	for _, nd := range ndList.Items {
		var ndName, ndIP string
		for _, addr := range nd.Status.Addresses {
			if addr.Type == v1.NodeHostName {
				ndName = addr.Address
			}
			if addr.Type == v1.NodeInternalIP {
				ndIP = addr.Address
			}
		}
		if ndName == hostName {
			return ndIP, ndList, nil
		}
	}
	return "", ndList, fmt.Errorf("k8s node list doesn't have ", hostName)
}