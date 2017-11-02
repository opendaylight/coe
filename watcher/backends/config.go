package backends

import (
	"k8s.io/client-go/kubernetes"
)

type Config struct {
	Backend   Coe
	ClientSet *kubernetes.Clientset
}
