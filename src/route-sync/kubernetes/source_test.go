package kubernetes_test

import (
	"route-sync/kubernetes"

	core "k8s.io/client-go/testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/runtime"
)

var _ = Describe("Source", func() {

	Context("TCP", func() {
		It("returns list of TCP routes", func() {
			nodeAddress := "10.0.0.0"
			nodePort := 42
			clientset := fake.NewSimpleClientset()
			clientset.PrependReactor("list", "services", func(action core.Action) (bool, runtime.Object, error) {
				port := v1.ServicePort{NodePort: int32(nodePort)}
				ports := []v1.ServicePort{port}
				spec := v1.ServiceSpec{Ports: ports}
				s1 := v1.Service{Spec: spec}
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

			endpoint := kubernetes.New(clientset)
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
			nodeAddress := "10.0.0.0"
			nodePort := 42
			clientset := fake.NewSimpleClientset()
			clientset.PrependReactor("list", "services", func(action core.Action) (bool, runtime.Object, error) {
				port := v1.ServicePort{NodePort: int32(nodePort)}
				ports := []v1.ServicePort{port}
				spec := v1.ServiceSpec{Ports: ports}
				s1 := v1.Service{Spec: spec}
				s2 := v1.Service{Spec: spec}
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

			endpoint := kubernetes.New(clientset)
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
			clientset := fake.NewSimpleClientset()
			clientset.PrependReactor("list", "services", func(action core.Action) (bool, runtime.Object, error) {
				serviceList := []v1.Service{}
				return true, &v1.ServiceList{Items: serviceList}, nil
			})
			clientset.PrependReactor("list", "nodes", func(action core.Action) (bool, runtime.Object, error) {
				nodesList := []v1.Node{}
				return true, &v1.NodeList{Items: nodesList}, nil
			})

			endpoint := kubernetes.New(clientset)
			routes, err := endpoint.TCP()
			Expect(err).To(BeNil())
			Expect(len(routes)).To(Equal(0))
		})
		It("returns only TCP routes for services with multiple ports", func() {
			nodeAddress := "10.0.0.0"
			udpNodePort := 42
			tcpNodePort := 43
			anotherTCPNodePort := 4342
			clientset := fake.NewSimpleClientset()
			clientset.PrependReactor("list", "services", func(action core.Action) (bool, runtime.Object, error) {
				port1 := v1.ServicePort{Protocol: "UDP", NodePort: int32(udpNodePort)}
				port2 := v1.ServicePort{Protocol: "TCP", NodePort: int32(tcpNodePort)}
				port3 := v1.ServicePort{Protocol: "TCP", NodePort: int32(anotherTCPNodePort)}
				ports := []v1.ServicePort{port1, port2, port3}
				spec := v1.ServiceSpec{Ports: ports}
				s1 := v1.Service{Spec: spec}
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

			endpoint := kubernetes.New(clientset)
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
			nodeAddress := "10.0.0.0"
			nodePort := 42
			clientset := fake.NewSimpleClientset()
			clientset.PrependReactor("list", "services", func(action core.Action) (bool, runtime.Object, error) {
				port1 := v1.ServicePort{Protocol: "TCP", NodePort: int32(nodePort)}
				port2 := v1.ServicePort{Protocol: "TCP"}
				ports := []v1.ServicePort{port1, port2}
				spec := v1.ServiceSpec{Ports: ports}
				s1 := v1.Service{Spec: spec}
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

			endpoint := kubernetes.New(clientset)
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

	})
})
