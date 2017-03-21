package kubernetes_test

import (
	"route-sync/kubernetes"
	"route-sync/route"
	"strconv"

	core "k8s.io/client-go/testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"
)

type k8sServiceData struct {
	serviceName string
	nodePort    int32
	labelName   string
	labelVal    string
}

func addServiceReactor(clientset *fake.Clientset, services []k8sServiceData) {
	clientset.PrependReactor("list", "services", func(action core.Action) (bool, runtime.Object, error) {
		serviceList := []v1.Service{}
		for _, serviceData := range services {
			ports := []v1.ServicePort{}
			if serviceData.nodePort != int32(-1) {
				port := v1.ServicePort{NodePort: int32(serviceData.nodePort)}
				ports = append(ports, port)
			}
			spec := v1.ServiceSpec{Ports: ports}
			labels := make(map[string]string)
			labels[serviceData.labelName] = serviceData.labelVal
			objectMeta := v1.ObjectMeta{Name: serviceData.serviceName, Labels: labels}
			s := v1.Service{Spec: spec, ObjectMeta: objectMeta}
			serviceList = append(serviceList, s)
		}
		return true, &v1.ServiceList{Items: serviceList}, nil
	})
}

var _ = Describe("Source", func() {
	const (
		domainName       = "cf.example.com"
		namespace        = "kubo-test"
		nodeAddress      = "10.0.0.0"
		nodePort         = route.Port(42)
		frontendPort     = route.Port(1000)
		serviceName      = "dashboard"
		anotherNamespace = "default"
		httpServiceLabel = "example-app"
	)
	var (
		clientset = fake.NewSimpleClientset()
	)

	Context("HTTP", func() {
		BeforeEach(func() {
			serviceData := k8sServiceData{
				serviceName: serviceName,
				nodePort:    int32(nodePort),
				labelName:   "http-route-sync",
				labelVal:    httpServiceLabel,
			}
			addServiceReactor(clientset, []k8sServiceData{serviceData})
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
		})
		It("returns list of HTTP routes", func() {
			endpoint := kubernetes.New(clientset, domainName)
			routes, err := endpoint.HTTP()
			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(1))
			httpRoute := routes[0]
			Expect(httpRoute.Name).To(Equal(httpServiceLabel + "." + domainName))
			backends := httpRoute.Backend
			Expect(len(backends)).To(Equal(1))
			Expect(backends[0].Port).To(Equal(nodePort))
			Expect(httpRoute.Backend[0].IP).To(Equal(nodeAddress))
		})
		It("returns list of HTTP routes across namespaces", func() {
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
			Expect(ns1Route.Name).To(Equal(httpServiceLabel + "." + domainName))
			backends := ns1Route.Backend
			Expect(len(backends)).To(Equal(1))
			Expect(backends[0].Port).To(Equal(nodePort))
			Expect(ns1Route.Backend[0].IP).To(Equal(nodeAddress))
			ns2Route := routes[1]
			Expect(ns2Route.Name).To(Equal(httpServiceLabel + "." + domainName))
		})

		It("skip HTTP routes without proper label", func() {
			serviceData := k8sServiceData{
				serviceName: serviceName,
				nodePort:    int32(nodePort),
				labelName:   "noRouteSync",
				labelVal:    "no",
			}
			addServiceReactor(clientset, []k8sServiceData{serviceData})
			endpoint := kubernetes.New(clientset, domainName)
			routes, err := endpoint.HTTP()
			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(0))
		})
	})

	Context("TCP", func() {
		BeforeEach(func() {
			serviceData := k8sServiceData{
				serviceName: serviceName,
				nodePort:    int32(nodePort),
				labelName:   "tcp-route-sync",
				labelVal:    "1000",
			}
			addServiceReactor(clientset, []k8sServiceData{serviceData})
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
		})
		It("returns list of TCP routes", func() {
			endpoint := kubernetes.New(clientset, "")
			routes, err := endpoint.TCP()
			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(1))
			tcpRoute := routes[0]
			Expect(tcpRoute.Frontend).To(Equal(frontendPort))
			backends := tcpRoute.Backend
			Expect(len(backends)).To(Equal(1))
			Expect(backends[0].Port).To(Equal(nodePort))
			Expect(tcpRoute.Backend[0].IP).To(Equal(nodeAddress))
		})
		It("returns list of TCP routes across namespaces", func() {
			clientset.PrependReactor("list", "namespaces", func(action core.Action) (bool, runtime.Object, error) {
				objectMeta := v1.ObjectMeta{Name: namespace}
				objectMeta2 := v1.ObjectMeta{Name: anotherNamespace}
				ns1 := v1.Namespace{ObjectMeta: objectMeta}
				ns2 := v1.Namespace{ObjectMeta: objectMeta2}
				namespaces := []v1.Namespace{ns1, ns2}
				return true, &v1.NamespaceList{Items: namespaces}, nil
			})
			endpoint := kubernetes.New(clientset, domainName)
			routes, err := endpoint.TCP()
			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(2))
			ns1Route := routes[0]
			Expect(ns1Route.Frontend).To(Equal(frontendPort))
			backends := ns1Route.Backend
			Expect(len(backends)).To(Equal(1))
			Expect(backends[0].Port).To(Equal(nodePort))
			Expect(ns1Route.Backend[0].IP).To(Equal(nodeAddress))
			ns2Route := routes[1]
			Expect(ns2Route.Frontend).To(Equal(frontendPort))
		})
		It("returns empty list of TCP routes when there are no services", func() {
			addServiceReactor(clientset, []k8sServiceData{})
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
				labels["tcp-route-sync"] = strconv.Itoa(int(frontendPort))
				objectMeta := v1.ObjectMeta{Labels: labels}
				s1 := v1.Service{ObjectMeta: objectMeta, Spec: spec}
				serviceList := []v1.Service{s1}
				return true, &v1.ServiceList{Items: serviceList}, nil
			})
			endpoint := kubernetes.New(clientset, "")
			routes, err := endpoint.TCP()
			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(2))
			tcpRoute := routes[0]
			Expect(tcpRoute.Frontend).To(Equal(frontendPort))
			backends := tcpRoute.Backend
			Expect(len(backends)).To(Equal(1))
			Expect(backends[0].Port).To(Equal(tcpNodePort))
			Expect(tcpRoute.Backend[0].IP).To(Equal(nodeAddress))
			Expect(routes[1].Backend[0].Port).To(Equal(anotherTCPNodePort))
		})
		It("skip routes for services with no NodePort", func() {
			serviceData := k8sServiceData{
				serviceName: serviceName,
				nodePort:    int32(-1),
				labelName:   "tcp-route-sync",
				labelVal:    "1000",
			}
			addServiceReactor(clientset, []k8sServiceData{serviceData})
			endpoint := kubernetes.New(clientset, "")
			routes, err := endpoint.TCP()
			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(0))
		})
		It("skip TCP routes for services without proper label", func() {
			serviceData := k8sServiceData{
				serviceName: serviceName,
				nodePort:    int32(-1),
				labelName:   "noRouteSync",
				labelVal:    "no",
			}
			addServiceReactor(clientset, []k8sServiceData{serviceData})
			endpoint := kubernetes.New(clientset, "")
			routes, err := endpoint.TCP()
			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(0))
		})

	})
})
