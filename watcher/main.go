package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/mitchellh/go-homedir"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig = flag.String("kubeconfig", "~/.kube/config", "path to kubernetes config file")
)

func main() {
	fmt.Println("Starting watcher")

	flag.Parse()

	kubeConfigFile, err := homedir.Expand(*kubeconfig)
	if err != nil {
		log.Fatalln(err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigFile)
	if err != nil {
		panic(err.Error())
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

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

	go watchPods(clientSet, wg, shutdown)
	go watchServices(clientSet, wg, shutdown)
	go watchEndpoints(clientSet, wg, shutdown)

	wg.Wait()
}

func watchPods(clientSet kubernetes.Interface, wg *sync.WaitGroup, shutdown <-chan struct{}) {
	informer := informers.NewSharedInformerFactory(clientSet, 10*time.Minute)
	podInformer := informer.Core().V1().Pods()
	podInformer.Informer().AddEventHandler(printPodWatcher{})
	podInformer.Informer().Run(shutdown)
	wg.Done()
}

func watchServices(clientSet kubernetes.Interface, wg *sync.WaitGroup, shutdown <-chan struct{}) {
	informer := informers.NewSharedInformerFactory(clientSet, 10*time.Minute)
	serviceInformer := informer.Core().V1().Services()
	serviceInformer.Informer().AddEventHandler(printServiceWatcher{})
	serviceInformer.Informer().Run(shutdown)
	wg.Done()
}

func watchEndpoints(clientSet kubernetes.Interface, wg *sync.WaitGroup, shutdown <-chan struct{}) {
	informer := informers.NewSharedInformerFactory(clientSet, 10*time.Minute)
	endpointInformer := informer.Core().V1().Endpoints()
	endpointInformer.Informer().AddEventHandler(printEndpointWatcher{})
	endpointInformer.Informer().Run(shutdown)
	wg.Done()
}
