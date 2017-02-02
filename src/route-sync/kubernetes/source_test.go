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
			Expect(tcpRoute.Frontend.Port).To(Equal(nodePort))
			backends := tcpRoute.Backend
			Expect(len(backends)).To(Equal(1))
			Expect(backends[0].Port).To(Equal(nodePort))
			Expect(tcpRoute.Backend[0].IP).To(Equal(nodeAddress))
		})
	})
})
