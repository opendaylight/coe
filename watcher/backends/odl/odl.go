package odl

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"k8s.io/client-go/pkg/api/v1"
	//"src/k8s.io/client-go/pkg/api/v1"

	"fmt"
	//"git.opendaylight.org/gerrit/p/coe.git/watcher/backends"
	"../../backends" // same project packages should be refereed by directory
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
	return b.putPod(string(pod.GetUID()), js)
}

func (b backend) UpdatePod(old, new *v1.Pod) error {
	return nil
}

func (b backend) DeletePod(pod *v1.Pod) error {
	b.deletePod(string(pod.GetUID()))
	return nil
}

func (b backend) AddNode(node *v1.Node) error {
	js, _ := json.MarshalIndent(node, "", "    ")
	js = createNodeStructure(node)
	return b.putNode(string(node.GetUID()), js)
}

func (b backend) UpdateNode(old, new *v1.Node) error {
	return nil
}

func (b backend) DeleteNode(node *v1.Node) error {
	return b.deleteNode(node.GetUID())
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
	return b.doRequest(http.MethodPut, b.urlPrefix + PodsUrl + uid, bytes.NewBuffer(js))
}

func (b backend) deletePod(uid string) error {
	return b.doRequest(http.MethodDelete, b.urlPrefix + PodsUrl + uid, nil)
}

func (b backend) putNode(uid string, js []byte) error {
	return b.doRequest(http.MethodPut, b.urlPrefix + NodesUrl + uid, bytes.NewBuffer(js))
}

func (b backend) deleteNode(uid string) error {
	return b.doRequest(http.MethodDelete, b.urlPrefix + NodesUrl + uid, nil)
}

func (b backend) putService(uid string, js []byte) error {
	return b.doRequest(http.MethodPut, b.urlPrefix + ServicesUrl + uid, bytes.NewBuffer(js))
}

func (b backend) deleteService(uid string) error {
	return b.doRequest(http.MethodDelete, b.urlPrefix + ServicesUrl + uid, nil)
}

func (b backend) putEndpoints(uid string, js []byte) error {
	return b.doRequest(http.MethodPut, b.urlPrefix + EndPointsUrl + uid, bytes.NewBuffer(js))
}

func (b backend) deleteEndpoints(uid string) error {
	return b.doRequest(http.MethodDelete, b.urlPrefix + EndPointsUrl + uid, nil)
}

func Watch(clientSet kubernetes.Interface, backend backends.Coe) {
	wg := &sync.WaitGroup{}

	wg.Add(4)

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
	go watchNodes(clientSet, wg, backend, shutdown)

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

func watchNodes(clientSet kubernetes.Interface, wg *sync.WaitGroup, backend backends.Coe, shutdown <-chan struct{}) {
	informer := informers.NewSharedInformerFactory(clientSet, 10*time.Minute)
	nodeInformer := informer.Core().V1().Nodes()
	nodeInformer.Informer().AddEventHandler(backends.NodesEventWatcher{Backend: backend})
	nodeInformer.Informer().Run(shutdown)
	wg.Done()
}