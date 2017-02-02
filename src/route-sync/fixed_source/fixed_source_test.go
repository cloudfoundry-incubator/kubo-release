package fixed_source_test

import (
	"route-sync/fixed_source"
	"route-sync/route"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FixedSource", func() {
	It("returns empty routes", func() {
		fs := fixed_source.New(nil, nil)

		tcpRoutes, err := fs.TCP()

		Expect(err).ShouldNot(HaveOccurred())
		Expect(tcpRoutes).To(BeEmpty())

		httpRoutes, err := fs.HTTP()

		Expect(err).ShouldNot(HaveOccurred())
		Expect(httpRoutes).To(BeEmpty())
	})

	It("returns TCP routes", func() {
		routes := []*route.TCP{&route.TCP{Frontend: route.Endpoint{IP: "10.0.0.1", Port: 8080},
			Backend: []route.Endpoint{route.Endpoint{IP: "10.10.10.10", Port: 9090}},
		}}

		fs := fixed_source.New(routes, nil)

		routes, error := fs.TCP()

		Expect(error).ShouldNot(HaveOccurred())
		Expect(routes).To(HaveLen(1))
	})

	It("returns HTTP routes", func() {
		routes := []*route.HTTP{&route.HTTP{Name: "foo.bar",
			Backend: []route.Endpoint{route.Endpoint{IP: "10.10.10.10", Port: 9090}},
		}}

		fs := fixed_source.New(nil, routes)

		routes, error := fs.HTTP()

		Expect(error).ShouldNot(HaveOccurred())
		Expect(routes).To(HaveLen(1))
	})
})
