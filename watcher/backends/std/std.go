package std

import (
	"encoding/json"
	"fmt"
	"log"

	"k8s.io/client-go/pkg/api/v1"
)

type Backend struct{}

func (b Backend) AddPod(pod *v1.Pod) error {
	fmt.Println("ADD:")
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
	fmt.Println("DELETE:")
	printJson(pod)
	return nil
}

func (b Backend) AddService(service *v1.Service) error {
	fmt.Println("DELETE:")
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
	fmt.Println("DELETE:")
	printJson(service)
	return nil
}

func (b Backend) AddEndpoints(endpoints *v1.Endpoints) error {
	fmt.Println("DELETE:")
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
	fmt.Println("DELETE:")
	printJson(endpoints)
	return nil
}

func printJson(obj interface{}) {
	b, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(b))
}
