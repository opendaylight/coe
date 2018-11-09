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
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var (
	EndpointsWatcher *endpointsWatcher
)

var endpointsStopCh chan struct{}

type EndpointsUpdate struct {
	Endpoints *api.Endpoints
	Op        utils.Operation
}

type endpointsWatcher struct {
	clientset           *kubernetes.Clientset
	endpointsController cache.Controller
	endpointsLister     cache.Indexer
	broadcaster         *utils.Broadcaster
}

type EndpointsUpdatesHandler interface {
	OnEndpointsUpdate(endpointsUpdate *EndpointsUpdate)
}

func (ew *endpointsWatcher) endpointsAddEventHandler(obj interface{}) {
	endpoints, ok := obj.(*api.Endpoints)
	if !ok {
		return
	}
	ew.broadcaster.Notify(&EndpointsUpdate{Op: utils.ADD, Endpoints: endpoints})
}

func (ew *endpointsWatcher) endpointsDeleteEventHandler(obj interface{}) {
	endpoints, ok := obj.(*api.Endpoints)
	if !ok {
		return
	}
	ew.broadcaster.Notify(&EndpointsUpdate{Op: utils.REMOVE, Endpoints: endpoints})
}

func (ew *endpointsWatcher) endpointsUpdateEventHandler(oldObj, newObj interface{}) {
	endpoints, ok := newObj.(*api.Endpoints)
	if !ok {
		return
	}
	if !reflect.DeepEqual(newObj, oldObj) {
		if endpoints.Name != "kube-scheduler" && endpoints.Name != "kube-controller-manager" {
			ew.broadcaster.Notify(&EndpointsUpdate{Op: utils.UPDATE, Endpoints: endpoints})
		}
	}
}

func (ew *endpointsWatcher) RegisterHandler(handler EndpointsUpdatesHandler) {
	ew.broadcaster.Add(utils.ListenerFunc(func(instance interface{}) {
		handler.OnEndpointsUpdate(instance.(*EndpointsUpdate))
	}))
}

func (ew *endpointsWatcher) List() []*api.Endpoints {
	objList := ew.endpointsLister.List()
	epInstances := make([]*api.Endpoints, len(objList))
	for i, ins := range objList {
		epInstances[i] = ins.(*api.Endpoints)
	}
	return epInstances
}

func (ew *endpointsWatcher) HasSynced() bool {
	return ew.endpointsController.HasSynced()
}

func StartEndpointsWatcher(clientset *kubernetes.Clientset, resyncPeriod time.Duration, namespace string, filter fields.Selector) (*endpointsWatcher, error) {

	ew := endpointsWatcher{}
	EndpointsWatcher = &ew

	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc:    ew.endpointsAddEventHandler,
		DeleteFunc: ew.endpointsDeleteEventHandler,
		UpdateFunc: ew.endpointsUpdateEventHandler,
	}

	ew.clientset = clientset
	ew.broadcaster = utils.NewBroadcaster()
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}
	if filter == nil {
		filter = fields.Everything()
	}
	lw := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "endpoints", namespace, filter)
	ew.endpointsLister, ew.endpointsController = cache.NewIndexerInformer(
		lw,
		&api.Endpoints{}, resyncPeriod, eventHandler,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	endpointsStopCh = make(chan struct{})
	go ew.endpointsController.Run(endpointsStopCh)
	return &ew, nil
}

func StopEndpointsWatcher() {
	endpointsStopCh <- struct{}{}
}
