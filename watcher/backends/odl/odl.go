package odl

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"k8s.io/client-go/pkg/api/v1"

	"git.opendaylight.org/gerrit/p/coe.git/watcher/backends"
)

func New(url string) backends.Coe {
	return backend{
		client:    &http.Client{},
		urlPrefix: url,
	}
}

type backend struct {
	client    *http.Client
	urlPrefix string
}

func (b backend) AddPod(pod *v1.Pod) error {
	js, _ := json.MarshalIndent(pod, "", "    ")
	js = createPodStructure(pod)
	b.putPod(string(pod.GetUID()), js)
	return nil
}

func (b backend) UpdatePod(old, new *v1.Pod) error {
	return nil
}

func (b backend) DeletePod(pod *v1.Pod) error {
	b.deletePod(string(pod.GetUID()))
	return nil
}

func createPodStructure(pod *v1.Pod) []byte {
	ipAddress := generateIP()
	interfaces := make([]Interface, 1)
	interfaces[0] = Interface{
		UID:            pod.GetUID(),
		NetworkID:      "00000000-0000-0000-0000-000000000000",
		IPAddress:      ipAddress,
		NetworkType:    "FLAT",
		SegmentationID: 0,
	}
	pods := make([]Pod, 1)
	pods[0] = Pod{
		UID:        pod.GetUID(),
		Interfaces: interfaces,
	}
	coe := Coe{
		Pods: pods,
	}
	js, _ := json.Marshal(coe)
	return js
}

func (b backend) AddService(service *v1.Service) error {
	return nil
}

func (b backend) UpdateService(old, new *v1.Service) error {
	return nil
}

func (b backend) DeleteService(service *v1.Service) error {
	return nil
}

func (b backend) AddEndpoints(endpoints *v1.Endpoints) error {
	return nil
}

func (b backend) UpdateEndpoints(old, new *v1.Endpoints) error {
	return nil
}

func (b backend) DeleteEndpoints(endpoints *v1.Endpoints) error {
	return nil
}

func (b backend) doRequest(method, url string, reader io.Reader) error {
	log.Println(method, url)
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("admin", "admin")

	res, err := b.client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Println(res)
		// TODO return an error
	}

	return nil
}

func (b backend) putPod(uid string, js []byte) error {
	return b.doRequest(http.MethodPut, b.urlPrefix+"/restconf/config/pod:coe/pods/"+uid, bytes.NewBuffer(js))
}

func (b backend) deletePod(uid string) error {
	return b.doRequest(http.MethodDelete, b.urlPrefix+"/restconf/config/pod:coe/pods/"+uid, nil)
}

func (b backend) putService(uid string, js []byte) error {
	return b.doRequest(http.MethodPut, b.urlPrefix+"/restconf/config/pod:coe/pods/"+uid, bytes.NewBuffer(js))
}

func (b backend) deleteService(uid string) error {
	return b.doRequest(http.MethodDelete, b.urlPrefix+"/restconf/config/pod:coe/pods/"+uid, nil)
}

func (b backend) putEndpoints(uid string, js []byte) error {
	return b.doRequest(http.MethodPut, b.urlPrefix+"/restconf/config/pod:coe/pods/"+uid, bytes.NewBuffer(js))
}

func (b backend) deleteEndpoints(uid string) error {
	return b.doRequest(http.MethodDelete, b.urlPrefix+"/restconf/config/pod:coe/pods/"+uid, nil)
}
