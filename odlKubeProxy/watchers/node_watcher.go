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
	NodeWatcher *nodeWatcher
)

var nodewatchStopCh chan struct{}

type NodeUpdate struct {
	Node *api.Node
	Op   utils.Operation
}

type nodeWatcher struct {
	clientset      *kubernetes.Clientset
	nodeController cache.Controller
	nodeLister     cache.Indexer
	broadcaster    *utils.Broadcaster
}

type NodeUpdatesHandler interface {
	OnNodeUpdate(nodeUpdate *NodeUpdate)
}

func (nw *nodeWatcher) nodeAddEventHandler(obj interface{}) {
	node, ok := obj.(*api.Node)
	if !ok {
		return
	}
	nw.broadcaster.Notify(&NodeUpdate{Op: utils.ADD, Node: node})
}

func (nw *nodeWatcher) nodeDeleteEventHandler(obj interface{}) {
	node, ok := obj.(*api.Node)
	if !ok {
		return
	}
	nw.broadcaster.Notify(&NodeUpdate{Op: utils.REMOVE, Node: node})
}

func (nw *nodeWatcher) nodeUpdateEventHandler(oldObj, newObj interface{}) {
	node, ok := newObj.(*api.Node)
	if !ok {
		return
	}
	if !reflect.DeepEqual(newObj, oldObj) {
		nw.broadcaster.Notify(&NodeUpdate{Op: utils.UPDATE, Node: node})
	}
}

func (nw *nodeWatcher) RegisterHandler(handler NodeUpdatesHandler) {
	nw.broadcaster.Add(utils.ListenerFunc(func(instance interface{}) {
		handler.OnNodeUpdate(instance.(*NodeUpdate))
	}))
}

func (nw *nodeWatcher) List() []*api.Node {
	objList := nw.nodeLister.List()
	nodeInstances := make([]*api.Node, len(objList))
	for i, ins := range objList {
		nodeInstances[i] = ins.(*api.Node)
	}
	return nodeInstances
}

func (nw *nodeWatcher) HasSynced() bool {
	return nw.nodeController.HasSynced()
}

func StartNodeWatcher(clientset *kubernetes.Clientset, resyncPeriod time.Duration, filter fields.Selector) (*nodeWatcher, error) {

	nw := nodeWatcher{}
	NodeWatcher = &nw
	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc:    nw.nodeAddEventHandler,
		DeleteFunc: nw.nodeDeleteEventHandler,
		UpdateFunc: nw.nodeUpdateEventHandler,
	}

	nw.clientset = clientset
	nw.broadcaster = utils.NewBroadcaster()

	if filter == nil {
		filter = fields.Everything()
	}

	lw := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "nodes", metav1.NamespaceAll, filter)
	nw.nodeLister, nw.nodeController = cache.NewIndexerInformer(
		lw,
		&api.Node{}, resyncPeriod, eventHandler,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	nodewatchStopCh = make(chan struct{})
	go nw.nodeController.Run(nodewatchStopCh)
	return &nw, nil
}

func StopNodeWatcher() {
	nodewatchStopCh <- struct{}{}
}
