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
			clientset := fake.NewSimpleClientset()
			clientset.AddReactor("list", "services", func(action core.Action) (bool, runtime.Object, error) {
				return true, &v1.ServiceList{}, nil
			})

			endpoint := kubernetes.New(clientset)
			routes, err := endpoint.TCP()
			Expect(err).To(BeNil())
			tcpRoute := routes[0]
			Expect(tcpRoute.Frontend.Port).To(Equal(42))
		})
	})
})
