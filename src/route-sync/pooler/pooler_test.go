package pooler_test

import (
	"context"
	"errors"
	"route-sync/pooler"
	"route-sync/route"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"

	"route-sync/route/routefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Time-based Pooler", func() {

	var (
		logger lager.Logger
		pool   pooler.Pooler
		src    *routefakes.FakeSource
		router *routefakes.FakeRouter
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("")
		pool = pooler.ByTime(time.Duration(0), logger)
		src = &routefakes.FakeSource{}
		tcpRoute := &route.TCP{Frontend: 8080,
			Backends: []route.Endpoint{route.Endpoint{IP: "10.10.0.10", Port: 9090}}}
		httpRoute := &route.HTTP{Name: "foo.bar.com",
			Backends: []route.Endpoint{route.Endpoint{IP: "10.10.0.10", Port: 9090}}}

		src.TCPReturns([]*route.TCP{tcpRoute}, nil)
		src.HTTPReturns([]*route.HTTP{httpRoute}, nil)
		router = &routefakes.FakeRouter{}
	})

	It("pools and passes", func() {
		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()
		go pool.Run(ctx, src, router)

		Eventually(func() bool {
			return router.HTTPCallCount() > 0 && router.TCPCallCount() > 0
		}).Should(BeTrue())
	})

	Context("with a failing TCP source", func() {
		BeforeEach(func() {
			src.TCPReturns(nil, errors.New("fail"))
		})

		It("panics", func() {
			Expect(func() { pool.Run(context.Background(), src, router) }).To(Panic())
			Expect(logger).To(gbytes.Say("fetching TCP routes"))
		})
	})

	Context("with a failing HTTP source", func() {
		BeforeEach(func() {
			src.HTTPReturns(nil, errors.New("fail"))
		})

		It("panics", func() {
			Expect(func() { pool.Run(context.Background(), src, router) }).To(Panic())
			Expect(logger).To(gbytes.Say("fetching HTTP routes"))
		})
	})

	Context("with a failing TCP router", func() {
		BeforeEach(func() {
			router.TCPReturns(errors.New("fail"))
		})

		It("panics", func() {
			Expect(func() { pool.Run(context.Background(), src, router) }).To(Panic())
			Expect(logger).To(gbytes.Say("posting TCP routes"))
		})
	})

	Context("with a failing HTTP router", func() {
		BeforeEach(func() {
			router.HTTPReturns(errors.New("fail"))
		})

		It("panics", func() {
			Expect(func() { pool.Run(context.Background(), src, router) }).To(Panic())
			Expect(logger).To(gbytes.Say("posting HTTP routes"))
		})
	})
})
