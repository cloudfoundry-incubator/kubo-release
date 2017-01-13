package pooler_test

import (
	"net"
	"route-sync/mocks"
	"route-sync/pooler"
	"route-sync/route"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Time-based Pooler", func() {

	It("starts and stops", func() {
		pool := pooler.ByTime(time.Duration(10))

		running := func() bool { return pool.Running() }

		done := pool.Start(&mocks.Source{}, &mocks.Announcer{})
		Eventually(running).Should(BeTrue())
		done <- struct{}{}
		Eventually(running).Should(BeFalse())
	})

	It("pools and passes", func() {
		pool := pooler.ByTime(time.Duration(10))

		src := &mocks.Source{}
		tcp_route := &route.TCP{Frontend: route.Endpoint{IP: net.ParseIP("10.0.0.1"), Port: 8080},
			Backend: []route.Endpoint{route.Endpoint{IP: net.ParseIP("10.10.0.10"), Port: 9090}}}

		src.TCP_value = []*route.TCP{tcp_route}
		announcer := &mocks.Announcer{}

		done := pool.Start(src, announcer)

		Eventually(func() bool { return src.TCP_count > 0 }).Should(BeTrue())
		Eventually(func() bool { return announcer.TCP_count > 0 }).Should(BeTrue())
		Expect(announcer.TCP_values[0][0]).To(Equal(tcp_route))

		done <- struct{}{}
	})
})
