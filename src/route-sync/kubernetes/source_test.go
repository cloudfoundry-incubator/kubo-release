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

type k8sServiceData struct {
	serviceName string
	udpPort     int32
	nodePort    int32
	labelName   string
	labelVal    string
}

func addServiceReactor(clientset *fake.Clientset, services []k8sServiceData) {
	clientset.PrependReactor("list", "services", func(action core.Action) (bool, runtime.Object, error) {
		serviceList := []v1.Service{}
		for _, serviceData := range services {
			ports := []v1.ServicePort{}
			if serviceData.nodePort > 0 {
				port := v1.ServicePort{NodePort: int32(serviceData.nodePort)}
				ports = append(ports, port)
			}
			if serviceData.udpPort > 0 {
				port := v1.ServicePort{Protocol: "UDP", NodePort: int32(serviceData.udpPort)}
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

func addNodeReactor(clientset *fake.Clientset, ipAddresses []string) {
	clientset.PrependReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		nodes := []v1.Node{}
		for _, ip := range ipAddresses {
			address := v1.NodeAddress{Type: "InternalIP", Address: ip}
			addresses := []v1.NodeAddress{address}
			ns := v1.NodeStatus{Addresses: addresses}
			node := v1.Node{Status: ns}
			nodes = append(nodes, node)
		}
		return true, &v1.NodeList{Items: nodes}, nil
	})
}

func addNamespaceReactor(clientset *fake.Clientset, namespaceNames []string) {
	clientset.PrependReactor("list", "namespaces", func(action core.Action) (bool, runtime.Object, error) {
		namespaces := []v1.Namespace{}
		for _, name := range namespaceNames {
			objectMeta := v1.ObjectMeta{Name: name}
			ns := v1.Namespace{ObjectMeta: objectMeta}
			namespaces = append(namespaces, ns)
		}
		return true, &v1.NamespaceList{Items: namespaces}, nil
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
		httpServiceLabel = "http-route-sync"
		httpRouteName    = "example-app"
		tcpServiceLabel  = "tcp-route-sync"
		tcpLabelPort     = "1000"
	)
	var (
		clientset = fake.NewSimpleClientset()
	)

	Context("HTTP", func() {
		It("returns list of HTTP routes", func() {
			serviceData := k8sServiceData{
				serviceName: serviceName,
				nodePort:    int32(nodePort),
				labelName:   httpServiceLabel,
				labelVal:    httpRouteName,
			}
			addServiceReactor(clientset, []k8sServiceData{serviceData})
			addNodeReactor(clientset, []string{nodeAddress})
			addNamespaceReactor(clientset, []string{namespace})

			endpoint := kubernetes.NewSource(clientset, domainName)
			routes, err := endpoint.HTTP()

			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(1))
			httpRoute := routes[0]
			Expect(httpRoute.Name).To(Equal(httpRouteName + "." + domainName))
			backends := httpRoute.Backends
			Expect(len(backends)).To(Equal(1))
			Expect(backends[0].Port).To(Equal(nodePort))
			Expect(httpRoute.Backends[0].IP).To(Equal(nodeAddress))
		})
		It("returns list multiple of HTTP routes", func() {
			serviceDataList := []k8sServiceData{}
			numServices := 10
			for i := 0; i < numServices; i++ {
				serviceData := k8sServiceData{
					serviceName: serviceName,
					nodePort:    int32(nodePort),
					labelName:   httpServiceLabel,
					labelVal:    httpRouteName,
				}
				serviceDataList = append(serviceDataList, serviceData)
			}
			addServiceReactor(clientset, serviceDataList)
			addNodeReactor(clientset, []string{nodeAddress})
			addNamespaceReactor(clientset, []string{namespace})

			endpoint := kubernetes.NewSource(clientset, domainName)
			routes, err := endpoint.HTTP()

			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(numServices))
		})
		It("returns list of HTTP routes across namespaces", func() {
			serviceData := k8sServiceData{
				serviceName: serviceName,
				nodePort:    int32(nodePort),
				labelName:   httpServiceLabel,
				labelVal:    httpRouteName,
			}
			addServiceReactor(clientset, []k8sServiceData{serviceData})
			addNodeReactor(clientset, []string{nodeAddress})
			addNamespaceReactor(clientset, []string{namespace, anotherNamespace})

			endpoint := kubernetes.NewSource(clientset, domainName)
			routes, err := endpoint.HTTP()

			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(2))
			ns1Route := routes[0]
			Expect(ns1Route.Name).To(Equal(httpRouteName + "." + domainName))
			backends := ns1Route.Backends
			Expect(len(backends)).To(Equal(1))
			Expect(backends[0].Port).To(Equal(nodePort))
			Expect(ns1Route.Backends[0].IP).To(Equal(nodeAddress))
			ns2Route := routes[1]
			Expect(ns2Route.Name).To(Equal(httpRouteName + "." + domainName))
		})

		It("skip HTTP routes without proper label", func() {
			serviceData := k8sServiceData{
				serviceName: serviceName,
				nodePort:    int32(nodePort),
				labelName:   "noRouteSync",
				labelVal:    "no",
			}
			addServiceReactor(clientset, []k8sServiceData{serviceData})
			addNodeReactor(clientset, []string{nodeAddress})
			addNamespaceReactor(clientset, []string{namespace})

			endpoint := kubernetes.NewSource(clientset, domainName)
			routes, err := endpoint.HTTP()

			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(0))
		})
	})

	Context("TCP", func() {
		It("returns list of TCP routes", func() {
			serviceData := k8sServiceData{
				serviceName: serviceName,
				nodePort:    int32(nodePort),
				labelName:   tcpServiceLabel,
				labelVal:    tcpLabelPort,
			}
			addServiceReactor(clientset, []k8sServiceData{serviceData})
			addNodeReactor(clientset, []string{nodeAddress})
			addNamespaceReactor(clientset, []string{namespace})

			endpoint := kubernetes.NewSource(clientset, "")
			routes, err := endpoint.TCP()

			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(1))
			tcpRoute := routes[0]
			Expect(tcpRoute.Frontend).To(Equal(frontendPort))
			backends := tcpRoute.Backends
			Expect(len(backends)).To(Equal(1))
			Expect(backends[0].Port).To(Equal(nodePort))
			Expect(tcpRoute.Backends[0].IP).To(Equal(nodeAddress))
		})
		It("returns list multiple of TCP routes", func() {
			serviceDataList := []k8sServiceData{}
			numServices := 10
			for i := 0; i < numServices; i++ {
				serviceData := k8sServiceData{
					serviceName: serviceName,
					nodePort:    int32(nodePort),
					labelName:   tcpServiceLabel,
					labelVal:    tcpLabelPort,
				}
				serviceDataList = append(serviceDataList, serviceData)
			}
			addServiceReactor(clientset, serviceDataList)
			addNodeReactor(clientset, []string{nodeAddress})
			addNamespaceReactor(clientset, []string{namespace})

			endpoint := kubernetes.NewSource(clientset, domainName)
			routes, err := endpoint.TCP()

			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(numServices))
		})
		It("returns list of TCP routes across namespaces", func() {
			serviceData := k8sServiceData{
				serviceName: serviceName,
				nodePort:    int32(nodePort),
				labelName:   tcpServiceLabel,
				labelVal:    tcpLabelPort,
			}
			addServiceReactor(clientset, []k8sServiceData{serviceData})
			addNodeReactor(clientset, []string{nodeAddress})
			addNamespaceReactor(clientset, []string{namespace, anotherNamespace})

			endpoint := kubernetes.NewSource(clientset, domainName)
			routes, err := endpoint.TCP()

			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(2))
			ns1Route := routes[0]
			Expect(ns1Route.Frontend).To(Equal(frontendPort))
			backends := ns1Route.Backends
			Expect(len(backends)).To(Equal(1))
			Expect(backends[0].Port).To(Equal(nodePort))
			Expect(ns1Route.Backends[0].IP).To(Equal(nodeAddress))
			ns2Route := routes[1]
			Expect(ns2Route.Frontend).To(Equal(frontendPort))
		})
		It("returns empty list of TCP routes when there are no services", func() {
			addServiceReactor(clientset, []k8sServiceData{})
			addNodeReactor(clientset, []string{nodeAddress})
			addNamespaceReactor(clientset, []string{namespace})

			endpoint := kubernetes.NewSource(clientset, "")
			routes, err := endpoint.TCP()

			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(0))
		})
		It("returns only TCP routes for services with multiple ports", func() {
			tcpPort := int32(43)
			serviceData := k8sServiceData{
				serviceName: serviceName,
				udpPort:     int32(42),
				nodePort:    tcpPort,
				labelName:   tcpServiceLabel,
				labelVal:    tcpLabelPort,
			}
			addServiceReactor(clientset, []k8sServiceData{serviceData})
			addNodeReactor(clientset, []string{nodeAddress})
			addNamespaceReactor(clientset, []string{namespace})

			endpoint := kubernetes.NewSource(clientset, "")
			routes, err := endpoint.TCP()

			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(1))
			tcpRoute := routes[0]
			Expect(tcpRoute.Frontend).To(Equal(frontendPort))
			backends := tcpRoute.Backends
			Expect(len(backends)).To(Equal(1))
			Expect(backends[0].Port).To(Equal(route.Port(tcpPort)))
			Expect(tcpRoute.Backends[0].IP).To(Equal(nodeAddress))
		})
		It("skip routes for services with no NodePort", func() {
			serviceData := k8sServiceData{
				serviceName: serviceName,
				nodePort:    int32(-1),
				labelName:   tcpServiceLabel,
				labelVal:    tcpLabelPort,
			}
			addServiceReactor(clientset, []k8sServiceData{serviceData})
			addNodeReactor(clientset, []string{nodeAddress})
			addNamespaceReactor(clientset, []string{namespace})

			endpoint := kubernetes.NewSource(clientset, "")
			routes, err := endpoint.TCP()

			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(0))
		})
		It("skip TCP routes for services without proper label", func() {
			addNodeReactor(clientset, []string{nodeAddress})
			addNamespaceReactor(clientset, []string{namespace})
			serviceData := k8sServiceData{
				serviceName: serviceName,
				nodePort:    int32(nodePort),
				labelName:   "noRouteSync",
				labelVal:    "no",
			}
			addServiceReactor(clientset, []k8sServiceData{serviceData})

			endpoint := kubernetes.NewSource(clientset, "")
			routes, err := endpoint.TCP()

			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(0))
		})

	})
})
