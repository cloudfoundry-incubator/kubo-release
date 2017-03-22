package fixedsource_test

import (
	"route-sync/fixedsource"
	"route-sync/route"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FixedSource", func() {
	It("returns empty routes", func() {
		fs := fixedsource.New(nil, nil)

		tcpRoutes, err := fs.TCP()

		Expect(err).ShouldNot(HaveOccurred())
		Expect(tcpRoutes).To(BeEmpty())

		httpRoutes, err := fs.HTTP()

		Expect(err).ShouldNot(HaveOccurred())
		Expect(httpRoutes).To(BeEmpty())
	})

	It("returns TCP routes", func() {
		routes := []*route.TCP{&route.TCP{Frontend: 8080,
			Backends: []route.Endpoint{route.Endpoint{IP: "10.10.10.10", Port: 9090}},
		}}

		fs := fixedsource.New(routes, nil)

		routes, error := fs.TCP()

		Expect(error).ShouldNot(HaveOccurred())
		Expect(routes).To(HaveLen(1))
	})

	It("returns HTTP routes", func() {
		routes := []*route.HTTP{&route.HTTP{Name: "foo.bar",
			Backends: []route.Endpoint{route.Endpoint{IP: "10.10.10.10", Port: 9090}},
		}}

		fs := fixedsource.New(nil, routes)

		routes, error := fs.HTTP()

		Expect(error).ShouldNot(HaveOccurred())
		Expect(routes).To(HaveLen(1))
	})
})
