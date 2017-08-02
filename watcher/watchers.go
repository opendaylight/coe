package main

import (
	"encoding/json"
	"fmt"
	"log"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
)

type Watchers struct {
	PodWatcher       cache.ResourceEventHandler
	ServiceWatcher   cache.ResourceEventHandler
	EndpointsWatcher cache.ResourceEventHandler
}

type printPodWatcher struct{}

func (watcher printPodWatcher) OnAdd(obj interface{}) {
	pod := obj.(*v1.Pod)
	fmt.Println("ADD Pod:", pod.GetUID(), pod.GetName(), pod.GetNamespace())
	printJson(pod)
}
func (watcher printPodWatcher) OnUpdate(oldObj, newObj interface{}) {
	//fmt.Println("UPDATE: ", oldObj, newObj)
}
func (watcher printPodWatcher) OnDelete(obj interface{}) {
	pod := obj.(*v1.Pod)
	fmt.Println("DELETE Pod:", pod.GetUID(), pod.GetName(), pod.GetNamespace())
	printJson(pod)
}

type printServiceWatcher struct{}

func (watcher printServiceWatcher) OnAdd(obj interface{}) {
	service := obj.(*v1.Service)
	fmt.Println("ADD Service:", service.GetUID(), service.GetName(), service.GetNamespace())
	printJson(service)
}
func (watcher printServiceWatcher) OnUpdate(oldObj, newObj interface{}) {
	//fmt.Println("UPDATE: ", oldObj, newObj)
}
func (watcher printServiceWatcher) OnDelete(obj interface{}) {
	service := obj.(*v1.Service)
	fmt.Println("DELETE Service:", service.GetUID(), service.GetName(), service.GetNamespace())
	printJson(service)
}

type printEndpointWatcher struct{}

func (watcher printEndpointWatcher) OnAdd(obj interface{}) {
	endpoint := obj.(*v1.Endpoints)
	fmt.Println("ADD Endpoint:", endpoint.GetUID(), endpoint.GetName(), endpoint.GetNamespace())
	printJson(endpoint)
}

func (watcher printEndpointWatcher) OnUpdate(oldObj, newObj interface{}) {
	//fmt.Println("UPDATE: ", oldObj, newObj)
}
func (watcher printEndpointWatcher) OnDelete(obj interface{}) {
	endpoint := obj.(*v1.Endpoints)
	fmt.Println("DELETE Endpoint:", endpoint.GetUID(), endpoint.GetName(), endpoint.GetNamespace())
	printJson(endpoint)
}

func printJson(obj interface{}) {
	b, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(b))
}
