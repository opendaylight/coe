package main

import (
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"

	"git.opendaylight.org/gerrit/p/coe.git/watcher/backends"
)

type Watchers struct {
	PodWatcher       cache.ResourceEventHandler
	ServiceWatcher   cache.ResourceEventHandler
	EndpointsWatcher cache.ResourceEventHandler
}

type podEventWatcher struct {
	backend backends.Coe
}

func (watcher podEventWatcher) OnAdd(obj interface{}) {
	pod := obj.(*v1.Pod)
	watcher.backend.AddPod(pod)
}
func (watcher podEventWatcher) OnUpdate(oldObj, newObj interface{}) {
	oldPod := oldObj.(*v1.Pod)
	newPod := newObj.(*v1.Pod)
	watcher.backend.UpdatePod(oldPod, newPod)
}
func (watcher podEventWatcher) OnDelete(obj interface{}) {
	pod := obj.(*v1.Pod)
	watcher.backend.DeletePod(pod)
}

type serviceEventWatcher struct {
	backend backends.Coe
}

func (watcher serviceEventWatcher) OnAdd(obj interface{}) {
	service := obj.(*v1.Service)
	watcher.backend.AddService(service)
}
func (watcher serviceEventWatcher) OnUpdate(oldObj, newObj interface{}) {
	oldService := oldObj.(*v1.Service)
	newService := newObj.(*v1.Service)
	watcher.backend.UpdateService(oldService, newService)
}
func (watcher serviceEventWatcher) OnDelete(obj interface{}) {
	service := obj.(*v1.Service)
	watcher.backend.DeleteService(service)
}

type endpointsEventWatcher struct {
	backend backends.Coe
}

func (watcher endpointsEventWatcher) OnAdd(obj interface{}) {
	endpoints := obj.(*v1.Endpoints)
	watcher.backend.AddEndpoints(endpoints)
}

func (watcher endpointsEventWatcher) OnUpdate(oldObj, newObj interface{}) {
	oldEndpoints := oldObj.(*v1.Endpoints)
	newEndpoints := newObj.(*v1.Endpoints)
	watcher.backend.UpdateEndpoints(oldEndpoints, newEndpoints)
}

func (watcher endpointsEventWatcher) OnDelete(obj interface{}) {
	endpoints := obj.(*v1.Endpoints)
	watcher.backend.DeleteEndpoints(endpoints)
}
