/*
 * Copyright (c) 2018 Kontron Canada Company and others.  All rights reserved.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v1.0 which accompanies this distribution,
 * and is available at http://www.eclipse.org/legal/epl-v10.html
 */

package watchers

import (
	"reflect"
	"time"

	"git.opendaylight.org/gerrit/p/coe.git/odlKubeProxy/utils"
	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var (
	ServiceWatcher *serviceWatcher
)

var servicesStopCh chan struct{}

type ServiceUpdate struct {
	Service *api.Service
	Op      utils.Operation
}

type serviceWatcher struct {
	clientset         kubernetes.Interface
	serviceController cache.Controller
	serviceLister     cache.Indexer
	broadcaster       *utils.Broadcaster
}

type ServiceUpdatesHandler interface {
	OnServiceUpdate(serviceUpdate *ServiceUpdate)
}

func (svcw *serviceWatcher) serviceAddEventHandler(obj interface{}) {
	service, ok := obj.(*api.Service)
	if !ok {
		return
	}
	svcw.broadcaster.Notify(&ServiceUpdate{Op: utils.ADD, Service: service})
}

func (svcw *serviceWatcher) serviceDeleteEventHandler(obj interface{}) {
	service, ok := obj.(*api.Service)
	if !ok {
		return
	}
	svcw.broadcaster.Notify(&ServiceUpdate{Op: utils.REMOVE, Service: service})
}

func (svcw *serviceWatcher) serviceUpdateEventHandler(oldObj, newObj interface{}) {
	service, ok := newObj.(*api.Service)
	if !ok {
		return
	}
	if !reflect.DeepEqual(newObj, oldObj) {
		svcw.broadcaster.Notify(&ServiceUpdate{Op: utils.UPDATE, Service: service})
	}
}

func (svcw *serviceWatcher) RegisterHandler(handler ServiceUpdatesHandler) {
	svcw.broadcaster.Add(utils.ListenerFunc(func(instance interface{}) {
		handler.OnServiceUpdate(instance.(*ServiceUpdate))
	}))
}

func (svcw *serviceWatcher) List() []*api.Service {
	objList := svcw.serviceLister.List()
	svcInstances := make([]*api.Service, len(objList))
	for i, ins := range objList {
		svcInstances[i] = ins.(*api.Service)
	}
	return svcInstances
}

func (svcw *serviceWatcher) HasSynced() bool {
	return svcw.serviceController.HasSynced()
}

func StartServiceWatcher(clientset kubernetes.Interface, resyncPeriod time.Duration, namespace string) (*serviceWatcher, error) {

	svcw := serviceWatcher{}
	ServiceWatcher = &svcw

	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc:    svcw.serviceAddEventHandler,
		DeleteFunc: svcw.serviceDeleteEventHandler,
		UpdateFunc: svcw.serviceUpdateEventHandler,
	}

	svcw.clientset = clientset
	svcw.broadcaster = utils.NewBroadcaster()

	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return clientset.CoreV1().Services(namespace).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return clientset.CoreV1().Services(namespace).Watch(options)
		},
	}

	svcw.serviceLister, svcw.serviceController = cache.NewIndexerInformer(
		lw,
		&api.Service{}, resyncPeriod, eventHandler,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	servicesStopCh = make(chan struct{})
	go svcw.serviceController.Run(servicesStopCh)
	return &svcw, nil
}

func StopServiceWatcher() {
	servicesStopCh <- struct{}{}
}
