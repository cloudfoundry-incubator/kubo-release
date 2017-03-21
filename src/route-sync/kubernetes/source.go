package kubernetes

import (
	"route-sync/route"
	"strconv"

	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

type endpoint struct {
	clientset k8s.Interface
	cfDomain  string
}

// New creates a route.Source for a given Kubernetes instance
func New(clientset k8s.Interface, cfDomain string) route.Source {
	return &endpoint{clientset: clientset, cfDomain: cfDomain}
}

func (e *endpoint) TCP() ([]*route.TCP, error) {
	namespaces, err := e.clientset.CoreV1().Namespaces().List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	ips, err := getIPs(e.clientset)
	if err != nil {
		return nil, err
	}

	routes := []*route.TCP{}
	for _, namespace := range namespaces.Items {
		services, err := e.clientset.CoreV1().Services(namespace.ObjectMeta.GetName()).List(v1.ListOptions{LabelSelector: "tcp-route-sync"})
		if err != nil {
			return nil, err
		}

		for _, service := range services.Items {
			for _, port := range service.Spec.Ports {
				if !isValidPort(port) {
					continue
				}
				portLabel, _ := strconv.Atoi(service.ObjectMeta.Labels["tcp-route-sync"])
				if portLabel == 0 {
					continue
				}
				frontendPort := route.Port(portLabel)
				nodePort := route.Port(port.NodePort)
				backends := getBackends(ips, nodePort)
				tcp := &route.TCP{Frontend: frontendPort, Backend: backends}
				routes = append(routes, tcp)
			}
		}
	}
	return routes, nil
}

func (e *endpoint) HTTP() ([]*route.HTTP, error) {
	namespaces, err := e.clientset.CoreV1().Namespaces().List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	ips, err := getIPs(e.clientset)
	if err != nil {
		return nil, err
	}

	routes := []*route.HTTP{}
	for _, namespace := range namespaces.Items {
		services, err := e.clientset.CoreV1().Services(namespace.ObjectMeta.GetName()).List(v1.ListOptions{LabelSelector: "http-route-sync"})
		if err != nil {
			return nil, err
		}

		for _, service := range services.Items {
			for _, port := range service.Spec.Ports {
				if !isValidPort(port) {
					continue
				}
				nodePort := route.Port(port.NodePort)
				backends := getBackends(ips, nodePort)
				routeName := service.ObjectMeta.Labels["http-route-sync"]
				fullName := routeName + "." + e.cfDomain
				http := &route.HTTP{Name: fullName, Backend: backends}
				routes = append(routes, http)
			}
		}
	}
	return routes, nil
}

// isValidPort returns true if this is a port we want to route to
func isValidPort(port v1.ServicePort) bool {
	if port.Protocol == "UDP" {
		return false
	}
	if route.Port(port.NodePort) <= 0 {
		return false
	}
	return true
}

// getIPs returns the IP of all minions
func getIPs(clientset k8s.Interface) ([]string, error) {
	nodes, err := clientset.CoreV1().Nodes().List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}
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

// getBackends returns a list of route.Endpoints for a set of backend IPs and a given nodePort
func getBackends(ips []string, nodePort route.Port) []route.Endpoint {
	backends := []route.Endpoint{}
	for _, ip := range ips {
		backends = append(backends, route.Endpoint{IP: ip, Port: nodePort})
	}
	return backends
}
