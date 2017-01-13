package fixed_source_test

import (
	"net"
	"route-sync/fixed_source"
	"route-sync/route"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FixedSource", func() {
	It("returns empty routes", func() {
		fs := fixed_source.New(nil)

		routes, error := fs.TCP()

		Expect(error).ShouldNot(HaveOccurred())
		Expect(routes).To(BeEmpty())
	})

	It("returns TCP routes", func() {
		routes := []*route.TCP{&route.TCP{Frontend: route.Endpoint{IP: net.ParseIP("10.0.0.1"), Port: 8080},
			Backend: []route.Endpoint{route.Endpoint{IP: net.ParseIP("10.10.10.10"), Port: 9090}},
		}}

		fs := fixed_source.New(routes)

		routes, error := fs.TCP()

		Expect(error).ShouldNot(HaveOccurred())
		Expect(routes).To(HaveLen(1))
	})
})
