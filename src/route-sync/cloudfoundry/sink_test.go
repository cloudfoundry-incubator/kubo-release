package cloudfoundry_test

import (
	. "route-sync/cloudfoundry"
	"route-sync/route"

	"code.cloudfoundry.org/route-registrar/messagebus/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sink", func() {
	Context("with a valid message bus", func() {
		var (
			sink route.Sink
			bus  fakes.FakeMessageBus
		)
		BeforeEach(func() {
			bus = fakes.FakeMessageBus{}
			sink = NewSink(&bus)
		})
		It("posts a single L7 route", func() {
			httpRoute := route.HTTP{
				Backend: []route.Endpoint{route.Endpoint{IP: "10.10.10.10", Port: 9090}},
				Name:    "foobar.cf-app.com",
			}
			sink.HTTP([]*route.HTTP{&httpRoute})
			Expect(bus.SendMessageCallCount()).To(Equal(1))
			subject, host, route, privateInstanceId := bus.SendMessageArgsForCall(0)

			Expect(subject).To(Equal("router.register"))
			Expect(host).To(Equal("10.10.10.10"))
			Expect(route.Port).To(Equal(9090))
			Expect(route.URIs).To(HaveLen(1))
			Expect(route.URIs[0]).To(Equal("foobar.cf-app.com"))
			Expect(privateInstanceId).NotTo(Equal(""))
		})
	})
})
