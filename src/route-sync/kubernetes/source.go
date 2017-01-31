package kubernetes

import "route-sync/route"

type endpoint struct {
}

func New(kubeconfig string) route.Source {
	return &endpoint{}
}

func (e *endpoint) TCP() ([]*route.TCP, error) {
	return []*route.TCP{}, nil
}
