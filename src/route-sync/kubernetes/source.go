package kubernetes

import (
	"route-sync/route"

	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

type endpoint struct {
	clientset k8s.Interface
}

func New(clientset k8s.Interface) route.Source {
	return &endpoint{clientset: clientset}
}

func (e *endpoint) TCP() ([]*route.TCP, error) {
	nodes, err := e.clientset.CoreV1().Nodes().List(v1.ListOptions{})

	if err != nil {
		panic(err)
	}

	services, err := e.clientset.CoreV1().Services("").List(v1.ListOptions{})

	if err != nil {
		panic(err)
	}

	ips := []string{}
	for _, node := range nodes.Items {
		for _, address := range node.Status.Addresses {
			if address.Type == "InternalIP" {
				ips = append(ips, address.Address)
			}
		}
	}

	routes := []*route.TCP{}
	for _, service := range services.Items {
		for _, port := range service.Spec.Ports {
			var nodePort int = int(port.NodePort)
			if port.Protocol == "UDP" {
				continue
			}
			if nodePort <= 0 {
				continue
			}
			backends := []route.Endpoint{}
			for _, ip := range ips {
				backends = append(backends, route.Endpoint{IP: ip, Port: nodePort})
			}

			tcp := &route.TCP{Frontend: nodePort, Backend: backends}
			routes = append(routes, tcp)
		}
	}
	return routes, nil
}

func (e *endpoint) HTTP() ([]*route.HTTP, error) {
	return []*route.HTTP{}, nil
}
