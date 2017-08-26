package backends

import (
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
)

type Watchers struct {
	PodWatcher       cache.ResourceEventHandler
	ServiceWatcher   cache.ResourceEventHandler
	EndpointsWatcher cache.ResourceEventHandler
}

type PodEventWatcher struct {
	Backend Coe
}

func (watcher PodEventWatcher) OnAdd(obj interface{}) {
	pod := obj.(*v1.Pod)
	watcher.Backend.AddPod(pod)
}
func (watcher PodEventWatcher) OnUpdate(oldObj, newObj interface{}) {
	oldPod := oldObj.(*v1.Pod)
	newPod := newObj.(*v1.Pod)
	watcher.Backend.UpdatePod(oldPod, newPod)
}
func (watcher PodEventWatcher) OnDelete(obj interface{}) {
	pod := obj.(*v1.Pod)
	watcher.Backend.DeletePod(pod)
}

type ServiceEventWatcher struct {
	Backend Coe
}

func (watcher ServiceEventWatcher) OnAdd(obj interface{}) {
	service := obj.(*v1.Service)
	watcher.Backend.AddService(service)
}
func (watcher ServiceEventWatcher) OnUpdate(oldObj, newObj interface{}) {
	oldService := oldObj.(*v1.Service)
	newService := newObj.(*v1.Service)
	watcher.Backend.UpdateService(oldService, newService)
}
func (watcher ServiceEventWatcher) OnDelete(obj interface{}) {
	service := obj.(*v1.Service)
	watcher.Backend.DeleteService(service)
}

type EndpointsEventWatcher struct {
	Backend Coe
}

func (watcher EndpointsEventWatcher) OnAdd(obj interface{}) {
	endpoints := obj.(*v1.Endpoints)
	watcher.Backend.AddEndpoints(endpoints)
}

func (watcher EndpointsEventWatcher) OnUpdate(oldObj, newObj interface{}) {
	oldEndpoints := oldObj.(*v1.Endpoints)
	newEndpoints := newObj.(*v1.Endpoints)
	watcher.Backend.UpdateEndpoints(oldEndpoints, newEndpoints)
}

func (watcher EndpointsEventWatcher) OnDelete(obj interface{}) {
	endpoints := obj.(*v1.Endpoints)
	watcher.Backend.DeleteEndpoints(endpoints)
}
