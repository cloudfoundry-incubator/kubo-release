package cloudfoundry_test

import (
	. "route-sync/cloudfoundry"
	"route-sync/route"

	messagebusfakes "code.cloudfoundry.org/route-registrar/messagebus/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sink", func() {
	Context("with a valid message bus", func() {
		var (
			sink       route.Sink
			msgBus     messagebusfakes.FakeMessageBus
			httpRoutes []*route.HTTP
		)
		BeforeEach(func() {
			msgBus = messagebusfakes.FakeMessageBus{}
			sink = NewSink(&msgBus)
			httpRoutes = []*route.HTTP{
				&route.HTTP{
					Backend: []route.Endpoint{route.Endpoint{IP: "10.10.10.10", Port: 9090}},
					Name:    "foobar.cf-app.com",
				},
				&route.HTTP{
					Backend: []route.Endpoint{
						route.Endpoint{IP: "10.10.10.10", Port: 9090},
						route.Endpoint{IP: "10.2.2.2", Port: 8080},
					},
					Name: "baz.cf-app.com",
				},
			}
		})
		It("posts a single L7 route", func() {
			sink.HTTP([]*route.HTTP{httpRoutes[0]})
			Expect(msgBus.SendMessageCallCount()).To(Equal(1))
			subject, host, route, privateInstanceId := msgBus.SendMessageArgsForCall(0)

			Expect(subject).To(Equal("router.register"))
			Expect(host).To(Equal("10.10.10.10"))
			Expect(route.Port).To(Equal(9090))
			Expect(route.URIs).To(HaveLen(1))
			Expect(route.URIs[0]).To(Equal("foobar.cf-app.com"))
			Expect(privateInstanceId).NotTo(Equal(""))
		})
		It("posts multiple L7 routes with multiple backends", func() {
			sink.HTTP(httpRoutes)
			Expect(msgBus.SendMessageCallCount()).To(Equal(3))
		})
		It("posts a sinlge L4 route", func() {

		})
	})
})
