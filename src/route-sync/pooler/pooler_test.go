package pooler_test

import (
	"route-sync/pooler"
	"route-sync/route"
	routefakes "route-sync/route/fakes"
	"time"

	"code.cloudfoundry.org/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Time-based Pooler", func() {

	It("starts and stops", func() {
		pool := pooler.ByTime(time.Duration(0), lager.NewLogger("pooler_test"))

		running := func() bool { return pool.Running() }

		done := pool.Start(&routefakes.Source{}, &routefakes.Router{})

		Eventually(running).Should(BeTrue())
		done <- struct{}{}
		Eventually(running).Should(BeFalse())
	})

	It("pools and passes", func() {
		pool := pooler.ByTime(time.Duration(0), lager.NewLogger("pooler_test"))

		src := &routefakes.Source{}
		tcpRoute := &route.TCP{Frontend: 8080,
			Backend: []route.Endpoint{route.Endpoint{IP: "10.10.0.10", Port: 9090}}}
		httpRoute := &route.HTTP{Name: "foo.bar.com",
			Backend: []route.Endpoint{route.Endpoint{IP: "10.10.0.10", Port: 9090}}}

		src.TCP_value = []*route.TCP{tcpRoute}
		src.HTTP_value = []*route.HTTP{httpRoute}
		router := &routefakes.Router{}

		done := pool.Start(src, router)

		Eventually(func() bool {
			src.Lock()
			defer src.Unlock()
			return src.TCP_count > 0
		}).Should(BeTrue())

		done <- struct{}{}

		Expect(router.TCP_count > 0).Should(BeTrue())
		Expect(router.TCP_values[0][0]).To(Equal(tcpRoute))

		Expect(src.HTTP_count > 0).Should(BeTrue())
		Expect(router.HTTP_count > 0).Should(BeTrue())
		Expect(router.HTTP_values[0][0]).To(Equal(httpRoute))
	})
})
