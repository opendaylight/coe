package backends

import (
	"k8s.io/client-go/pkg/api/v1"
)

type Coe interface {
	AddPod(*v1.Pod) error
	UpdatePod(old, new *v1.Pod) error
	DeletePod(*v1.Pod) error

	AddService(*v1.Service) error
	UpdateService(old, new *v1.Service) error
	DeleteService(*v1.Service) error

	AddEndpoints(*v1.Endpoints) error
	UpdateEndpoints(old, new *v1.Endpoints) error
	DeleteEndpoints(*v1.Endpoints) error

	AddNode(*v1.Node) error
	UpdateNode(old, new *v1.Node) error
	DeleteNode(*v1.Node) error
}
