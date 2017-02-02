package kubernetes

import (
	"route-sync/route"

	k8s "k8s.io/client-go/kubernetes"
)

type endpoint struct {
}

func New(clientset k8s.Interface) route.Source {
	return &endpoint{}
}

func (e *endpoint) TCP() ([]*route.TCP, error) {
	return []*route.TCP{}, nil
}

func (e *endpoint) HTTP() ([]*route.HTTP, error) {
	return []*route.HTTP{}, nil
}
