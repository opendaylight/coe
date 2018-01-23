package odl

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"k8s.io/client-go/pkg/api/v1"

	"git.opendaylight.org/gerrit/p/coe.git/watcher/backends"
)

type backend struct {
	client    *http.Client
	urlPrefix string
	username  string
	password  string
}

func New(url, username, password string) backends.Coe {
	return backend{
		client:    &http.Client{},
		username:  username,
		password:  password,
		urlPrefix: url,
	}
}

func (b backend) AddPod(pod *v1.Pod) error {
	js := createPodStructure(pod)
	return b.putPod(string(pod.GetUID()), js)
}

func (b backend) UpdatePod(old, new *v1.Pod) error {
	newJs := createPodStructure(new)
	return b.putPod(string(new.GetUID()), newJs)
}

func (b backend) DeletePod(pod *v1.Pod) error {
	return b.deletePod(string(pod.GetUID()))
}

func (b backend) AddNode(node *v1.Node) error {
	js := createNodeStructure(node)
	return b.putNode(string(node.GetUID()), js)
}

func (b backend) UpdateNode(old, new *v1.Node) error {
	newJs := createNodeStructure(new)
	return b.putNode(string(new.GetUID()), newJs)
}

func (b backend) DeleteNode(node *v1.Node) error {
	return b.deleteNode(string(node.GetUID()))
}

func (b backend) AddService(service *v1.Service) error {
	js := createServiceStructure(service)
	return b.putService(string(service.GetUID()), js)
}

func (b backend) UpdateService(old, new *v1.Service) error {
	newJs := createServiceStructure(new)
	return b.putService(string(new.GetUID()), newJs)
}

func (b backend) DeleteService(service *v1.Service) error {
	return b.deleteService(string(service.GetUID()))
}

func (b backend) AddEndpoints(endpoints *v1.Endpoints) error {
	js := createEndpointStructure(endpoints)
	return b.putEndpoints(string(endpoints.GetUID()), js)
}

func (b backend) UpdateEndpoints(old, new *v1.Endpoints) error {
	newJs := createEndpointStructure(new)
	log.Println(newJs)
	return b.putEndpoints(string(new.GetUID()), newJs)
}

func (b backend) DeleteEndpoints(endpoints *v1.Endpoints) error {
	return b.deleteEndpoints(string(endpoints.GetUID()))
}

func (b backend) doRequest(method, url string, reader io.Reader) error {
	log.Println(method, url)
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(b.username, b.password)

	res, err := b.client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Println(res)
		return fmt.Errorf("HTTP server did not respond with 200 OK: %v", res)
	}

	return nil
}

func (b backend) putPod(uid string, js []byte) error {
	return b.doRequest(http.MethodPut, b.urlPrefix+PodsUrl+uid, bytes.NewBuffer(js))
}

func (b backend) deletePod(uid string) error {
	return b.doRequest(http.MethodDelete, b.urlPrefix+PodsUrl+uid, nil)
}

func (b backend) putNode(uid string, js []byte) error {
	return b.doRequest(http.MethodPut, b.urlPrefix+NodesUrl+uid, bytes.NewBuffer(js))
}

func (b backend) deleteNode(uid string) error {
	return b.doRequest(http.MethodDelete, b.urlPrefix+NodesUrl+uid, nil)
}

func (b backend) putService(uid string, js []byte) error {
	return b.doRequest(http.MethodPut, b.urlPrefix+ServicesUrl+uid, bytes.NewBuffer(js))
}

func (b backend) deleteService(uid string) error {
	return b.doRequest(http.MethodDelete, b.urlPrefix+ServicesUrl+uid, nil)
}

func (b backend) putEndpoints(uid string, js []byte) error {
	return b.doRequest(http.MethodPut, b.urlPrefix+EndPointsUrl+uid, bytes.NewBuffer(js))
}

func (b backend) deleteEndpoints(uid string) error {
	return b.doRequest(http.MethodDelete, b.urlPrefix+EndPointsUrl+uid, nil)
}
