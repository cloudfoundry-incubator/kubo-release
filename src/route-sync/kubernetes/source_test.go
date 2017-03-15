package kubernetes_test

import (
	"route-sync/kubernetes"
	"route-sync/route"

	core "k8s.io/client-go/testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"
)

var _ = Describe("Source", func() {
	const (
		domainName       = "cf.example.com"
		namespace        = "kubo-test"
		nodeAddress      = "10.0.0.0"
		nodePort         = route.Port(42)
		serviceName      = "dashboard"
		anotherNamespace = "default"
	)
	var (
		clientset = fake.NewSimpleClientset()
	)

	Context("HTTP", func() {
		It("returns list of HTTP routes", func() {
			clientset.PrependReactor("list", "services", func(action core.Action) (bool, runtime.Object, error) {
				port := v1.ServicePort{NodePort: int32(nodePort)}
				ports := []v1.ServicePort{port}
				spec := v1.ServiceSpec{Ports: ports}
				labels := make(map[string]string)
				labels["http-route-sync"] = "true"
				objectMeta := v1.ObjectMeta{Name: serviceName, Labels: labels}
				s1 := v1.Service{Spec: spec, ObjectMeta: objectMeta}
				serviceList := []v1.Service{s1}
				return true, &v1.ServiceList{Items: serviceList}, nil
			})
			clientset.PrependReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
				address := v1.NodeAddress{Type: "InternalIP", Address: nodeAddress}
				addresses := []v1.NodeAddress{address}
				ns := v1.NodeStatus{Addresses: addresses}
				n1 := v1.Node{Status: ns}
				nodesList := []v1.Node{n1}
				return true, &v1.NodeList{Items: nodesList}, nil
			})
			clientset.PrependReactor("list", "namespaces", func(action core.Action) (bool, runtime.Object, error) {
				objectMeta := v1.ObjectMeta{Name: namespace}
				ns1 := v1.Namespace{ObjectMeta: objectMeta}
				namespaces := []v1.Namespace{ns1}
				return true, &v1.NamespaceList{Items: namespaces}, nil
			})
			endpoint := kubernetes.New(clientset, domainName)
			routes, err := endpoint.HTTP()
			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(1))
			httpRoute := routes[0]
			Expect(httpRoute.Name).To(Equal(serviceName + "." + namespace + "." + domainName))
			backends := httpRoute.Backend
			Expect(len(backends)).To(Equal(1))
			Expect(backends[0].Port).To(Equal(nodePort))
			Expect(httpRoute.Backend[0].IP).To(Equal(nodeAddress))
		})
		It("returns list of HTTP routes across namespaces", func() {
			clientset.PrependReactor("list", "services", func(action core.Action) (bool, runtime.Object, error) {
				port := v1.ServicePort{NodePort: int32(nodePort)}
				ports := []v1.ServicePort{port}
				spec := v1.ServiceSpec{Ports: ports}
				labels := make(map[string]string)
				labels["http-route-sync"] = "true"
				objectMeta := v1.ObjectMeta{Name: serviceName, Labels: labels}
				s1 := v1.Service{Spec: spec, ObjectMeta: objectMeta}
				serviceList := []v1.Service{s1}
				return true, &v1.ServiceList{Items: serviceList}, nil
			})
			clientset.PrependReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
				address := v1.NodeAddress{Type: "InternalIP", Address: nodeAddress}
				addresses := []v1.NodeAddress{address}
				ns := v1.NodeStatus{Addresses: addresses}
				n1 := v1.Node{Status: ns}
				nodesList := []v1.Node{n1}
				return true, &v1.NodeList{Items: nodesList}, nil
			})
			clientset.PrependReactor("list", "namespaces", func(action core.Action) (bool, runtime.Object, error) {
				objectMeta := v1.ObjectMeta{Name: namespace}
				objectMeta2 := v1.ObjectMeta{Name: anotherNamespace}
				ns1 := v1.Namespace{ObjectMeta: objectMeta}
				ns2 := v1.Namespace{ObjectMeta: objectMeta2}
				namespaces := []v1.Namespace{ns1, ns2}
				return true, &v1.NamespaceList{Items: namespaces}, nil
			})
			endpoint := kubernetes.New(clientset, domainName)
			routes, err := endpoint.HTTP()
			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(2))
			ns1Route := routes[0]
			Expect(ns1Route.Name).To(Equal(serviceName + "." + namespace + "." + domainName))
			backends := ns1Route.Backend
			Expect(len(backends)).To(Equal(1))
			Expect(backends[0].Port).To(Equal(nodePort))
			Expect(ns1Route.Backend[0].IP).To(Equal(nodeAddress))
			ns2Route := routes[1]
			Expect(ns2Route.Name).To(Equal(serviceName + "." + anotherNamespace + "." + domainName))
		})

		It("skip HTTP routes without proper label", func() {
			clientset.PrependReactor("list", "services", func(action core.Action) (bool, runtime.Object, error) {
				port := v1.ServicePort{NodePort: int32(nodePort)}
				ports := []v1.ServicePort{port}
				spec := v1.ServiceSpec{Ports: ports}
				objectMeta := v1.ObjectMeta{Name: serviceName}
				s1 := v1.Service{Spec: spec, ObjectMeta: objectMeta}
				labels := make(map[string]string)
				labels["noRouteSync"] = "true"
				objectMeta2 := v1.ObjectMeta{Name: serviceName, Labels: labels}
				s2 := v1.Service{Spec: spec, ObjectMeta: objectMeta2}
				serviceList := []v1.Service{s1, s2}
				return true, &v1.ServiceList{Items: serviceList}, nil
			})
			clientset.PrependReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
				address := v1.NodeAddress{Type: "InternalIP", Address: nodeAddress}
				addresses := []v1.NodeAddress{address}
				ns := v1.NodeStatus{Addresses: addresses}
				n1 := v1.Node{Status: ns}
				nodesList := []v1.Node{n1}
				return true, &v1.NodeList{Items: nodesList}, nil
			})
			clientset.PrependReactor("list", "namespaces", func(action core.Action) (bool, runtime.Object, error) {
				objectMeta := v1.ObjectMeta{Name: namespace}
				objectMeta2 := v1.ObjectMeta{Name: anotherNamespace}
				ns1 := v1.Namespace{ObjectMeta: objectMeta}
				ns2 := v1.Namespace{ObjectMeta: objectMeta2}
				namespaces := []v1.Namespace{ns1, ns2}
				return true, &v1.NamespaceList{Items: namespaces}, nil
			})
			endpoint := kubernetes.New(clientset, domainName)
			routes, err := endpoint.HTTP()
			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(0))
		})
	})

	Context("TCP", func() {
		It("returns list of TCP routes", func() {
			clientset.PrependReactor("list", "services", func(action core.Action) (bool, runtime.Object, error) {
				port := v1.ServicePort{NodePort: int32(nodePort)}
				ports := []v1.ServicePort{port}
				spec := v1.ServiceSpec{Ports: ports}
				labels := make(map[string]string)
				labels["tcp-route-sync"] = "true"
				objectMeta := v1.ObjectMeta{Labels: labels}
				s1 := v1.Service{ObjectMeta: objectMeta, Spec: spec}
				serviceList := []v1.Service{s1}
				return true, &v1.ServiceList{Items: serviceList}, nil
			})
			clientset.PrependReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
				address := v1.NodeAddress{Type: "InternalIP", Address: nodeAddress}
				addresses := []v1.NodeAddress{address}
				ns := v1.NodeStatus{Addresses: addresses}
				n1 := v1.Node{Status: ns}
				nodesList := []v1.Node{n1}
				return true, &v1.NodeList{Items: nodesList}, nil
			})
			clientset.PrependReactor("list", "namespaces", func(action core.Action) (bool, runtime.Object, error) {
				objectMeta := v1.ObjectMeta{Name: namespace}
				ns1 := v1.Namespace{ObjectMeta: objectMeta}
				namespaces := []v1.Namespace{ns1}
				return true, &v1.NamespaceList{Items: namespaces}, nil
			})

			endpoint := kubernetes.New(clientset, "")
			routes, err := endpoint.TCP()
			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(1))
			tcpRoute := routes[0]
			Expect(tcpRoute.Frontend).To(Equal(nodePort))
			backends := tcpRoute.Backend
			Expect(len(backends)).To(Equal(1))
			Expect(backends[0].Port).To(Equal(nodePort))
			Expect(tcpRoute.Backend[0].IP).To(Equal(nodeAddress))
		})
		It("returns list of TCP routes", func() {
			clientset.PrependReactor("list", "services", func(action core.Action) (bool, runtime.Object, error) {
				port := v1.ServicePort{NodePort: int32(nodePort)}
				ports := []v1.ServicePort{port}
				spec := v1.ServiceSpec{Ports: ports}
				labels := make(map[string]string)
				labels["tcp-route-sync"] = "true"
				objectMeta := v1.ObjectMeta{Labels: labels}
				s1 := v1.Service{ObjectMeta: objectMeta, Spec: spec}
				s2 := v1.Service{ObjectMeta: objectMeta, Spec: spec}
				serviceList := []v1.Service{s1, s2}
				return true, &v1.ServiceList{Items: serviceList}, nil
			})
			clientset.PrependReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
				address := v1.NodeAddress{Type: "InternalIP", Address: nodeAddress}
				addresses := []v1.NodeAddress{address}
				ns := v1.NodeStatus{Addresses: addresses}
				n1 := v1.Node{Status: ns}
				n2 := v1.Node{Status: ns}
				nodesList := []v1.Node{n1, n2}
				return true, &v1.NodeList{Items: nodesList}, nil
			})
			clientset.PrependReactor("list", "namespaces", func(action core.Action) (bool, runtime.Object, error) {
				objectMeta := v1.ObjectMeta{Name: namespace}
				ns1 := v1.Namespace{ObjectMeta: objectMeta}
				namespaces := []v1.Namespace{ns1}
				return true, &v1.NamespaceList{Items: namespaces}, nil
			})

			endpoint := kubernetes.New(clientset, "")
			routes, err := endpoint.TCP()
			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(2))
			tcpRoute := routes[0]
			Expect(tcpRoute.Frontend).To(Equal(nodePort))
			backends := tcpRoute.Backend
			Expect(len(backends)).To(Equal(2))
			Expect(backends[0].Port).To(Equal(nodePort))
			Expect(tcpRoute.Backend[0].IP).To(Equal(nodeAddress))
		})
		It("returns empty list of TCP routes when there are no services", func() {
			clientset.PrependReactor("list", "services", func(action core.Action) (bool, runtime.Object, error) {
				serviceList := []v1.Service{}
				return true, &v1.ServiceList{Items: serviceList}, nil
			})
			clientset.PrependReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
				nodesList := []v1.Node{}
				return true, &v1.NodeList{Items: nodesList}, nil
			})

			endpoint := kubernetes.New(clientset, "")
			routes, err := endpoint.TCP()
			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(0))
		})
		It("returns only TCP routes for services with multiple ports", func() {
			udpNodePort := route.Port(42)
			tcpNodePort := route.Port(43)
			anotherTCPNodePort := route.Port(4342)
			clientset.PrependReactor("list", "services", func(action core.Action) (bool, runtime.Object, error) {
				port1 := v1.ServicePort{Protocol: "UDP", NodePort: int32(udpNodePort)}
				port2 := v1.ServicePort{Protocol: "TCP", NodePort: int32(tcpNodePort)}
				port3 := v1.ServicePort{Protocol: "TCP", NodePort: int32(anotherTCPNodePort)}
				ports := []v1.ServicePort{port1, port2, port3}
				spec := v1.ServiceSpec{Ports: ports}
				labels := make(map[string]string)
				labels["tcp-route-sync"] = "true"
				objectMeta := v1.ObjectMeta{Labels: labels}
				s1 := v1.Service{ObjectMeta: objectMeta, Spec: spec}
				serviceList := []v1.Service{s1}
				return true, &v1.ServiceList{Items: serviceList}, nil
			})
			clientset.PrependReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
				address := v1.NodeAddress{Type: "InternalIP", Address: nodeAddress}
				addresses := []v1.NodeAddress{address}
				ns := v1.NodeStatus{Addresses: addresses}
				n1 := v1.Node{Status: ns}
				nodesList := []v1.Node{n1}
				return true, &v1.NodeList{Items: nodesList}, nil
			})
			clientset.PrependReactor("list", "namespaces", func(action core.Action) (bool, runtime.Object, error) {
				objectMeta := v1.ObjectMeta{Name: namespace}
				ns1 := v1.Namespace{ObjectMeta: objectMeta}
				namespaces := []v1.Namespace{ns1}
				return true, &v1.NamespaceList{Items: namespaces}, nil
			})

			endpoint := kubernetes.New(clientset, "")
			routes, err := endpoint.TCP()
			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(2))
			tcpRoute := routes[0]
			Expect(tcpRoute.Frontend).To(Equal(tcpNodePort))
			backends := tcpRoute.Backend
			Expect(len(backends)).To(Equal(1))
			Expect(backends[0].Port).To(Equal(tcpNodePort))
			Expect(tcpRoute.Backend[0].IP).To(Equal(nodeAddress))
			Expect(routes[1].Backend[0].Port).To(Equal(anotherTCPNodePort))
		})
		It("skip routes for services with no NodePort", func() {
			clientset.PrependReactor("list", "services", func(action core.Action) (bool, runtime.Object, error) {
				port1 := v1.ServicePort{Protocol: "TCP", NodePort: int32(nodePort)}
				port2 := v1.ServicePort{Protocol: "TCP"}
				ports := []v1.ServicePort{port1, port2}
				spec := v1.ServiceSpec{Ports: ports}
				labels := make(map[string]string)
				labels["tcp-route-sync"] = "true"
				objectMeta := v1.ObjectMeta{Labels: labels}
				s1 := v1.Service{ObjectMeta: objectMeta, Spec: spec}
				serviceList := []v1.Service{s1}
				return true, &v1.ServiceList{Items: serviceList}, nil
			})
			clientset.PrependReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
				address := v1.NodeAddress{Type: "InternalIP", Address: nodeAddress}
				addresses := []v1.NodeAddress{address}
				ns := v1.NodeStatus{Addresses: addresses}
				n1 := v1.Node{Status: ns}
				nodesList := []v1.Node{n1}
				return true, &v1.NodeList{Items: nodesList}, nil
			})
			clientset.PrependReactor("list", "namespaces", func(action core.Action) (bool, runtime.Object, error) {
				objectMeta := v1.ObjectMeta{Name: namespace}
				ns1 := v1.Namespace{ObjectMeta: objectMeta}
				namespaces := []v1.Namespace{ns1}
				return true, &v1.NamespaceList{Items: namespaces}, nil
			})

			endpoint := kubernetes.New(clientset, "")
			routes, err := endpoint.TCP()
			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(1))
			tcpRoute := routes[0]
			Expect(tcpRoute.Frontend).To(Equal(nodePort))
			backends := tcpRoute.Backend
			Expect(len(backends)).To(Equal(1))
			Expect(backends[0].Port).To(Equal(nodePort))
			Expect(tcpRoute.Backend[0].IP).To(Equal(nodeAddress))
		})
		It("skip TCP routes for services without proper label", func() {
			clientset.PrependReactor("list", "services", func(action core.Action) (bool, runtime.Object, error) {
				port1 := v1.ServicePort{Protocol: "TCP", NodePort: int32(nodePort)}
				port2 := v1.ServicePort{Protocol: "TCP"}
				ports := []v1.ServicePort{port1, port2}
				spec := v1.ServiceSpec{Ports: ports}
				labels := make(map[string]string)
				labels["noRouteSync"] = "true"
				objectMeta := v1.ObjectMeta{Labels: labels}
				s1 := v1.Service{ObjectMeta: objectMeta, Spec: spec}
				s2 := v1.Service{Spec: spec}
				serviceList := []v1.Service{s1, s2}
				return true, &v1.ServiceList{Items: serviceList}, nil
			})
			clientset.PrependReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
				address := v1.NodeAddress{Type: "InternalIP", Address: nodeAddress}
				addresses := []v1.NodeAddress{address}
				ns := v1.NodeStatus{Addresses: addresses}
				n1 := v1.Node{Status: ns}
				nodesList := []v1.Node{n1}
				return true, &v1.NodeList{Items: nodesList}, nil
			})
			clientset.PrependReactor("list", "namespaces", func(action core.Action) (bool, runtime.Object, error) {
				objectMeta := v1.ObjectMeta{Name: namespace}
				ns1 := v1.Namespace{ObjectMeta: objectMeta}
				namespaces := []v1.Namespace{ns1}
				return true, &v1.NamespaceList{Items: namespaces}, nil
			})

			endpoint := kubernetes.New(clientset, "")
			routes, err := endpoint.TCP()
			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(0))
		})

	})
})
