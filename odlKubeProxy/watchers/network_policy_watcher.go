/*
 * Copyright (c) 2018 Kontron Canada Company and others.  All rights reserved.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v1.0 which accompanies this distribution,
 * and is available at http://www.eclipse.org/legal/epl-v10.html
 */

package watchers

import (
	"errors"
	"reflect"
	"strconv"
	"time"

	"git.opendaylight.org/gerrit/p/coe.git/odlKubeProxy/utils"
	apiextensions "k8s.io/api/extensions/v1beta1"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var (
	NetworkPolicyWatcher *networkPolicyWatcher
)

var networkPolicyStopCh chan struct{}

type NetworkPolicyUpdate struct {
	NetworkPolicy interface{}
	Op            utils.Operation
}

type networkPolicyWatcher struct {
	clientset               *kubernetes.Clientset
	networkPolicyController cache.Controller
	networkPolicyLister     cache.Indexer
	broadcaster             *utils.Broadcaster
}

type NetworkPolicyUpdatesHandler interface {
	OnNetworkPolicyUpdate(networkPolicyUpdate *NetworkPolicyUpdate)
}

func (npw *networkPolicyWatcher) networkPolicyAddEventHandler(obj interface{}) {
	npw.broadcaster.Notify(&NetworkPolicyUpdate{Op: utils.ADD, NetworkPolicy: obj})
}

func (npw *networkPolicyWatcher) networkPolicyDeleteEventHandler(obj interface{}) {
	npw.broadcaster.Notify(&NetworkPolicyUpdate{Op: utils.REMOVE, NetworkPolicy: obj})
}

func (npw *networkPolicyWatcher) networkPolicyUpdateEventHandler(oldObj, newObj interface{}) {
	if !reflect.DeepEqual(newObj, oldObj) {
		npw.broadcaster.Notify(&NetworkPolicyUpdate{Op: utils.UPDATE, NetworkPolicy: newObj})
	}
}

func (npw *networkPolicyWatcher) RegisterHandler(handler NetworkPolicyUpdatesHandler) {
	npw.broadcaster.Add(utils.ListenerFunc(func(instance interface{}) {
		handler.OnNetworkPolicyUpdate(instance.(*NetworkPolicyUpdate))
	}))
}

func (npw *networkPolicyWatcher) List() []interface{} {
	return npw.networkPolicyLister.List()
}

func (npw *networkPolicyWatcher) HasSynced() bool {
	return npw.networkPolicyController.HasSynced()
}

func StartNetworkPolicyWatcher(clientset *kubernetes.Clientset, resyncPeriod time.Duration, namespace string, filter fields.Selector) (*networkPolicyWatcher, error) {

	npw := networkPolicyWatcher{}
	NetworkPolicyWatcher = &npw

	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc:    npw.networkPolicyAddEventHandler,
		DeleteFunc: npw.networkPolicyDeleteEventHandler,
		UpdateFunc: npw.networkPolicyUpdateEventHandler,
	}

	npw.clientset = clientset

	v1NetworkPolicy := true
	v, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, errors.New("Failed to get API server version due to " + err.Error())
	}

	//Just in case but we shouldn't support k8s less than v1.7
	minorVer, _ := strconv.Atoi(v.Minor)
	if v.Major == "1" && minorVer < 7 {
		v1NetworkPolicy = false
	}

	if namespace == "" {
		namespace = metav1.NamespaceAll
	}
	if filter == nil {
		filter = fields.Everything()
	}

	npw.broadcaster = utils.NewBroadcaster()
	var lw *cache.ListWatch
	if v1NetworkPolicy {
		lw = cache.NewListWatchFromClient(clientset.NetworkingV1().RESTClient(), "networkpolicies", namespace, filter)
		npw.networkPolicyLister, npw.networkPolicyController = cache.NewIndexerInformer(
			lw, &networking.NetworkPolicy{}, resyncPeriod, eventHandler,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
		)
	} else {
		lw = cache.NewListWatchFromClient(clientset.ExtensionsV1beta1().RESTClient(), "networkpolicies", namespace, filter)
		npw.networkPolicyLister, npw.networkPolicyController = cache.NewIndexerInformer(
			lw, &apiextensions.NetworkPolicy{}, resyncPeriod, eventHandler,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
		)
	}
	networkPolicyStopCh = make(chan struct{})
	go npw.networkPolicyController.Run(networkPolicyStopCh)
	return &npw, nil
}

func StopNetworkPolicyWatcher() {
	networkPolicyStopCh <- struct{}{}
}
