package std

import (
	"encoding/json"
	"fmt"
	"log"

	"k8s.io/client-go/pkg/api/v1"
)

type Backend struct{}

func (b Backend) AddPod(pod *v1.Pod) error {
	fmt.Println("Add:")
	printJson(pod)
	return nil
}

func (b Backend) UpdatePod(old, new *v1.Pod) error {
	fmt.Println("Update:")
	fmt.Println("Old:")
	printJson(old)
	fmt.Println("New:")
	printJson(new)
	return nil
}

func (b Backend) DeletePod(pod *v1.Pod) error {
	fmt.Println("Delete:")
	printJson(pod)
	return nil
}

func (b Backend) AddService(service *v1.Service) error {
	fmt.Println("Add:")
	printJson(service)
	return nil
}

func (b Backend) UpdateService(old, new *v1.Service) error {
	fmt.Println("Update:")
	fmt.Println("Old:")
	printJson(old)
	fmt.Println("New:")
	printJson(new)
	return nil
}

func (b Backend) DeleteService(service *v1.Service) error {
	fmt.Println("Delete:")
	printJson(service)
	return nil
}

func (b Backend) AddEndpoints(endpoints *v1.Endpoints) error {
	fmt.Println("Add:")
	printJson(endpoints)
	return nil
}

func (b Backend) UpdateEndpoints(old, new *v1.Endpoints) error {
	fmt.Println("Update:")
	fmt.Println("Old:")
	printJson(old)
	fmt.Println("New:")
	printJson(new)
	return nil
}

func (b Backend) DeleteEndpoints(endpoints *v1.Endpoints) error {
	fmt.Println("Delete:")
	printJson(endpoints)
	return nil
}

func (b Backend) AddNode(node *v1.Node) error {
	fmt.Println("Add:")
	printJson(node)
	return nil
}

func (b Backend) UpdateNode(old, new *v1.Node) error {
	fmt.Println("Update:")
	fmt.Println("Old:")
	printJson(old)
	fmt.Println("New:")
	printJson(new)
	return nil
}

func (b Backend) DeleteNode(node *v1.Node) error {
	fmt.Println("Delete:")
	printJson(node)
	return nil
}

func printJson(obj interface{}) {
	b, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(b))
}
