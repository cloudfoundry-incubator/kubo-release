package kubernetes

import (
	"route-sync/route"
        "k8s.io/client-go/pkg/api/v1"
	k8s "k8s.io/client-go/kubernetes"
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
		if len(service.Spec.Ports) != 1 {
			continue
		}
		var nodePort int = int(service.Spec.Ports[0].NodePort)
		frontend := nodePort

		backends := []route.Endpoint{}
		for _, ip := range ips {
			backends = append(backends, route.Endpoint{IP: ip, Port: nodePort})
		}

		tcp := &route.TCP{Frontend: frontend, Backend: backends}
		routes = append(routes, tcp)
        }
	return routes, nil
}

func (e *endpoint) HTTP() ([]*route.HTTP, error) {
	return []*route.HTTP{}, nil
}
