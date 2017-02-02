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

	namespaces, err := e.clientset.CoreV1().Namespaces().List(v1.ListOptions{})
	if err != nil {
		panic(err)
	}

	routes := []*route.TCP{}
	for _, namespace := range namespaces.Items {
		services, err := e.clientset.CoreV1().Services(namespace.ObjectMeta.GetName()).List(v1.ListOptions{})
		if err != nil {
			panic(err)
		}

		ips, _ := GetIPs(nodes)

		for _, service := range services.Items {
			for _, port := range service.Spec.Ports {
				var nodePort int = int(port.NodePort)
				if port.Protocol == "UDP" {
					continue
				}
				if nodePort <= 0 {
					continue
				}
				backends, _ := GetBackends(ips, nodePort)
				tcp := &route.TCP{Frontend: nodePort, Backend: backends}
				routes = append(routes, tcp)
			}
		}
	}
	return routes, nil
}

func (e *endpoint) HTTP() ([]*route.HTTP, error) {
	nodes, err := e.clientset.CoreV1().Nodes().List(v1.ListOptions{})
	if err != nil {
		panic(err)
	}

	namespaces, err := e.clientset.CoreV1().Namespaces().List(v1.ListOptions{})
	if err != nil {
		panic(err)
	}

	routes := []*route.HTTP{}
	for _, namespace := range namespaces.Items {
		services, err := e.clientset.CoreV1().Services(namespace.ObjectMeta.GetName()).List(v1.ListOptions{})
		if err != nil {
			panic(err)
		}

		ips, _ := GetIPs(nodes)

		for _, service := range services.Items {
			for _, port := range service.Spec.Ports {
				// TODO Which ports should we include for HTTP routing?
				var nodePort int = int(port.NodePort)
				if port.Protocol == "UDP" {
					continue
				}
				if nodePort <= 0 {
					continue
				}
				backends, _ := GetBackends(ips, nodePort)
				// TODO Append a CF domain onto the Name param
				http := &route.HTTP{Name: service.ObjectMeta.GetName() + "." + namespace.ObjectMeta.GetName(), Backend: backends}
				routes = append(routes, http)
			}
		}
	}
	return routes, nil
}

func GetIPs(nodes *v1.NodeList) ([]string, error) {
	ips := []string{}
	for _, node := range nodes.Items {
		for _, address := range node.Status.Addresses {
			if address.Type == "InternalIP" {
				ips = append(ips, address.Address)
			}
		}
	}
	return ips, nil
}

func GetBackends(ips []string, nodePort int) ([]route.Endpoint, error) {
	backends := []route.Endpoint{}
	for _, ip := range ips {
		backends = append(backends, route.Endpoint{IP: ip, Port: nodePort})
	}
	return backends, nil
}
