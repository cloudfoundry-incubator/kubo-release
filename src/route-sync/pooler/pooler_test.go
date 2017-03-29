package pooler_test

import (
	"context"
	"route-sync/pooler"
	"route-sync/route"
	"time"

	"code.cloudfoundry.org/lager"

	"route-sync/route/routefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Time-based Pooler", func() {

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

		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()
		go pool.Run(ctx, src, router)

		Eventually(func() bool {
			return router.HTTPCallCount() > 0 && router.TCPCallCount() > 0
		}).Should(BeTrue())
	})
})
