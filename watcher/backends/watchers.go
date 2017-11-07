package backends

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	syncTime = 10 * time.Minute
)

type Watchers struct {
	PodWatcher       cache.ResourceEventHandler
	NodeWatcher      cache.ResourceEventHandler
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
	if isPodUpdated(oldPod, newPod) {
		watcher.Backend.UpdatePod(oldPod, newPod)
	}
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
	if isServiceUpdated(oldService, newService) {
		watcher.Backend.UpdateService(oldService, newService)
	}
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
	if isEndpointsUpdated(oldEndpoints, newEndpoints) {
		watcher.Backend.UpdateEndpoints(oldEndpoints, newEndpoints)
	}
}

func (watcher EndpointsEventWatcher) OnDelete(obj interface{}) {
	endpoints := obj.(*v1.Endpoints)
	watcher.Backend.DeleteEndpoints(endpoints)
}

type NodesEventWatcher struct {
	Backend Coe
}

func (watcher NodesEventWatcher) OnAdd(obj interface{}) {
	node := obj.(*v1.Node)
	watcher.Backend.AddNode(node)
}

func (watcher NodesEventWatcher) OnUpdate(oldObj, newObj interface{}) {
	oldNode := oldObj.(*v1.Node)
	newNode := newObj.(*v1.Node)
	if isNodeUpdated(oldNode, newNode) {
		watcher.Backend.UpdateNode(oldNode, newNode)
	}
}

func (watcher NodesEventWatcher) OnDelete(obj interface{}) {
	node := obj.(*v1.Node)
	watcher.Backend.DeleteNode(node)
}

func Watch(clientSet kubernetes.Interface, backend Coe) {
	wg := &sync.WaitGroup{}

	wg.Add(4)

	shutdown := make(chan struct{})

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	go func() {
		for range signalChannel {
			fmt.Println()
			fmt.Println("Shutting down")
			close(shutdown)
			break
		}
	}()

	informer := informers.NewSharedInformerFactory(clientSet, syncTime)

	// We use typedInformer.Run(shutdown) which blocks until the informer is properly shut down.
	// informer.Start() does not block and we have no way of ensuring informers have properly
	// shut down.
	go watchPods(informer, wg, backend, shutdown)
	go watchNodes(informer, wg, backend, shutdown)
	go watchServices(informer, wg, backend, shutdown)
	go watchEndpoints(informer, wg, backend, shutdown)

	wg.Wait()
}

func watchPods(informer informers.SharedInformerFactory, wg *sync.WaitGroup, backend Coe, shutdown <-chan struct{}) {
	podInformer := informer.Core().V1().Pods()
	podInformer.Informer().AddEventHandler(PodEventWatcher{Backend: backend})
	podInformer.Informer().Run(shutdown)
	wg.Done()
}

func watchServices(informer informers.SharedInformerFactory, wg *sync.WaitGroup, backend Coe, shutdown <-chan struct{}) {
	serviceInformer := informer.Core().V1().Services()
	serviceInformer.Informer().AddEventHandler(ServiceEventWatcher{Backend: backend})
	serviceInformer.Informer().Run(shutdown)
	wg.Done()
}

func watchEndpoints(informer informers.SharedInformerFactory, wg *sync.WaitGroup, backend Coe, shutdown <-chan struct{}) {
	endpointInformer := informer.Core().V1().Endpoints()
	endpointInformer.Informer().AddEventHandler(EndpointsEventWatcher{Backend: backend})
	endpointInformer.Informer().Run(shutdown)
	wg.Done()
}

func watchNodes(informer informers.SharedInformerFactory, wg *sync.WaitGroup, backend Coe, shutdown <-chan struct{}) {
	nodeInformer := informer.Core().V1().Nodes()
	nodeInformer.Informer().AddEventHandler(NodesEventWatcher{Backend: backend})
	nodeInformer.Informer().Run(shutdown)
	wg.Done()
}
