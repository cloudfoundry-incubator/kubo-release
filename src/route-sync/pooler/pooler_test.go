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
		logger     lager.Logger
		pool       pooler.Pooler
		src        *routefakes.FakeSource
		router     *routefakes.FakeRouter
		tcpRoutes  []*route.TCP
		httpRoutes []*route.HTTP
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("")
		pool = pooler.ByTime(time.Duration(0), logger)
		src = &routefakes.FakeSource{}

		tcpRoutes = []*route.TCP{
			&route.TCP{
				Frontend: 9999999,
				Backends: []route.Endpoint{
					{IP: "TCP Backend", Port: 11111},
				},
			},
		}

		httpRoutes = []*route.HTTP{
			&route.HTTP{
				Name: "HTTP Frontend",
				Backends: []route.Endpoint{
					{IP: "HTTP Backend", Port: 22222},
				},
			},
		}

		src.TCPReturns(tcpRoutes, nil)
		src.HTTPReturns(httpRoutes, nil)
		router = &routefakes.FakeRouter{}
	})

	It("passes routes from the source to the router", func() {
		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		go pool.Run(ctx, src, router)

		Eventually(func() bool {
			return router.HTTPCallCount() > 0 && router.TCPCallCount() > 0
		}).Should(BeTrue())

		httpRoutesRecieved := router.HTTPArgsForCall(0)
		tcpRoutesRecieved := router.TCPArgsForCall(0)

		Expect(httpRoutesRecieved).Should(Equal(httpRoutes))
		Expect(tcpRoutesRecieved).Should(Equal(tcpRoutes))

		Expect(logger).To(gbytes.Say("registered routes"))
		Expect(logger).To(gbytes.Say("HTTP Frontend"))
		Expect(logger).To(gbytes.Say("HTTP Backend"))
		Expect(logger).To(gbytes.Say("TCP Backend"))
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
