package odl

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"k8s.io/client-go/pkg/api/v1"

	"fmt"
	"git.opendaylight.org/gerrit/p/coe.git/watcher/backends"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"os"
	"os/signal"
	"sync"
	"time"
)

func New(url, username, password string) backends.Coe {
	return backend{
		client:    &http.Client{},
		username:  username,
		password:  password,
		urlPrefix: url,
	}
}

type backend struct {
	client    *http.Client
	urlPrefix string
	username  string
	password  string
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
	req.SetBasicAuth(b.username, b.password)

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

func Watch(clientSet kubernetes.Interface, backend backends.Coe) {
	wg := &sync.WaitGroup{}

	wg.Add(3)

	shutdown := make(chan struct{})

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	go func() {
		for range signalChannel {
			fmt.Println()
			fmt.Println("Shutting down")
			close(shutdown)
			break
		}
	}()

	go watchPods(clientSet, wg, backend, shutdown)
	go watchServices(clientSet, wg, backend, shutdown)
	go watchEndpoints(clientSet, wg, backend, shutdown)

	wg.Wait()
}

func watchPods(clientSet kubernetes.Interface, wg *sync.WaitGroup, backend backends.Coe, shutdown <-chan struct{}) {
	informer := informers.NewSharedInformerFactory(clientSet, 10*time.Minute)
	podInformer := informer.Core().V1().Pods()
	podInformer.Informer().AddEventHandler(backends.PodEventWatcher{Backend: backend})
	podInformer.Informer().Run(shutdown)
	wg.Done()
}

func watchServices(clientSet kubernetes.Interface, wg *sync.WaitGroup, backend backends.Coe, shutdown <-chan struct{}) {
	informer := informers.NewSharedInformerFactory(clientSet, 10*time.Minute)
	serviceInformer := informer.Core().V1().Services()
	serviceInformer.Informer().AddEventHandler(backends.ServiceEventWatcher{Backend: backend})
	serviceInformer.Informer().Run(shutdown)
	wg.Done()
}

func watchEndpoints(clientSet kubernetes.Interface, wg *sync.WaitGroup, backend backends.Coe, shutdown <-chan struct{}) {
	informer := informers.NewSharedInformerFactory(clientSet, 10*time.Minute)
	endpointInformer := informer.Core().V1().Endpoints()
	endpointInformer.Informer().AddEventHandler(backends.EndpointsEventWatcher{Backend: backend})
	endpointInformer.Informer().Run(shutdown)
	wg.Done()
}
