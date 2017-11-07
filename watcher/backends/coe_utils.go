package backends

import (
	"reflect"

	"k8s.io/client-go/pkg/api/v1"
)

func isNodeUpdated(oldNode *v1.Node, newNode *v1.Node) bool {
	if oldNode.Spec.PodCIDR != newNode.Spec.PodCIDR {
		return true
	}
	return !reflect.DeepEqual(oldNode.Status.Addresses, newNode.Status.Addresses)
}

func isPodUpdated(oldPod *v1.Pod, newPod *v1.Pod) bool {
	if oldPod.Status.PodIP != newPod.Status.PodIP {
		return true
	}
	if oldPod.Status.HostIP != newPod.Status.HostIP {
		return true
	}
	if oldPod.GetName() != newPod.GetName() {
		return false
	}
	if oldPod.GetNamespace() != newPod.GetNamespace() {
		return true
	}
	if oldPod.GetUID() != newPod.GetUID() {
		return true
	}
	return false
}

func isServiceUpdated(oldService *v1.Service, newService *v1.Service) bool {
	if !reflect.DeepEqual(oldService.Spec.Ports, newService.Spec.Ports) {
		return true
	}
	if !reflect.DeepEqual(oldService.Spec.ExternalIPs, newService.Spec.ExternalIPs) {
		return true
	}
	if !reflect.DeepEqual(oldService.Status.LoadBalancer.Ingress, newService.Status.LoadBalancer.Ingress) {
		return true
	}
	if oldService.Spec.ClusterIP != newService.Spec.ClusterIP {
		return true
	}
	if oldService.Spec.LoadBalancerIP != newService.Spec.LoadBalancerIP {
		return true
	}
	if oldService.GetName() != newService.GetName() {
		return true
	}
	if oldService.GetNamespace() != newService.GetNamespace() {
		return true
	}
	return false
}

func isEndpointsUpdated(oldEndpoints *v1.Endpoints, newEndpoints *v1.Endpoints) bool {
	if len(oldEndpoints.Subsets) != len(newEndpoints.Subsets) {
		return true
	}
	for i := 0; i < len(oldEndpoints.Subsets); i++ {
		if !reflect.DeepEqual(oldEndpoints.Subsets[i].Addresses, newEndpoints.Subsets[i].Addresses) {
			return true
		}
		if !reflect.DeepEqual(oldEndpoints.Subsets[i].Ports, newEndpoints.Subsets[i].Ports) {
			return true
		}
	}
	if oldEndpoints.GetNamespace() != newEndpoints.GetNamespace() {
		return true
	}
	if oldEndpoints.GetName() != newEndpoints.GetName() {
		return true
	}
	return false
}
