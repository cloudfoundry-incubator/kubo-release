package pooler_test

import (
	"route-sync/pooler"
	"route-sync/route"
	"time"

	"code.cloudfoundry.org/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"route-sync/route/routefakes"
)

var _ = Describe("Time-based Pooler", func() {

	It("starts and stops", func() {
		pool := pooler.ByTime(time.Duration(0), lager.NewLogger("pooler_test"))

		running := func() bool { return pool.Running() }

		done := pool.Start(&routefakes.FakeSource{}, &routefakes.FakeRouter{})



		Eventually(running).Should(BeTrue())
		done <- struct{}{}
		Eventually(running).Should(BeFalse())
	})

	It("pools and passes", func() {
		pool := pooler.ByTime(time.Duration(0), lager.NewLogger("pooler_test"))

		src := &routefakes.FakeSource{}
		tcpRoute := &route.TCP{Frontend: 8080,
			Backends: []route.Endpoint{route.Endpoint{IP: "10.10.0.10", Port: 9090}}}
		httpRoute := &route.HTTP{Name: "foo.bar.com",
			Backends: []route.Endpoint{route.Endpoint{IP: "10.10.0.10", Port: 9090}}}

		src.TCPReturns([]*route.TCP{tcpRoute}, nil)
		src.HTTPReturns([]*route.HTTP{httpRoute}, nil)
		router := &routefakes.FakeRouter{}


		done := pool.Start(src, router)

		Eventually(func() bool {
			return router.HTTPCallCount() > 0 && router.TCPCallCount() > 0
		}).Should(BeTrue())

		done <- struct{}{}


		Expect(src.HTTPCallCount() > 0).Should(BeTrue())
		Expect(src.TCPCallCount() > 0).Should(BeTrue())
	})
})
