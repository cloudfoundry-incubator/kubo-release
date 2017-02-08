package cloudfoundry_test

import (
	. "route-sync/cloudfoundry"
	"route-sync/cloudfoundry/tcp"
	tcpfakes "route-sync/cloudfoundry/tcp/fakes"
	"route-sync/route"

	messagebusfakes "code.cloudfoundry.org/route-registrar/messagebus/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Router", func() {
	Context("with a valid message bus", func() {
		var (
			router     route.Router
			tcpRouter  tcpfakes.FakeRouter
			msgBus     messagebusfakes.FakeMessageBus
			httpRoutes []*route.HTTP
			tcpRoutes  []*route.TCP
		)
		BeforeEach(func() {
			msgBus = messagebusfakes.FakeMessageBus{}
			tcpRouter = tcpfakes.FakeRouter{}
			tcpRouter.RouterGroupsResult = []tcp.RouterGroup{tcp.RouterGroup{
				Guid:            "abc123",
				Name:            "myRouter",
				ReservablePorts: "1000-2000",
				Type:            "tcp",
			},
			}
			router = NewRouter(&msgBus, &tcpRouter)
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

			tcpRoutes = []*route.TCP{
				&route.TCP{
					Backend:  []route.Endpoint{route.Endpoint{IP: "10.10.10.10", Port: 9090}},
					Frontend: 1010,
				},
				&route.TCP{
					Backend: []route.Endpoint{
						route.Endpoint{IP: "10.10.10.10", Port: 9090},
						route.Endpoint{IP: "10.2.2.2", Port: 8080},
					},
					Frontend: 2020,
				},
			}
		})
		It("posts a single L7 route", func() {
			Expect(router.HTTP([]*route.HTTP{httpRoutes[0]})).To(Succeed())

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
			Expect(router.HTTP(httpRoutes)).To(Succeed())
			Expect(msgBus.SendMessageCallCount()).To(Equal(3))
		})
		It("posts a sinlge L4 route", func() {
			Expect(router.TCP([]*route.TCP{tcpRoutes[0]})).To(Succeed())

			Expect(tcpRouter.CreateRoutesLastRoutes).To(ConsistOf(tcpRoutes[0]))
			Expect(tcpRouter.CreateRoutesLastRouterGroup).To(Equal(tcpRouter.RouterGroupsResult[0]))
		})
		It("posts multiple L4 routes", func() {
			Expect(router.TCP(tcpRoutes)).To(Succeed())

			Expect(tcpRouter.CreateRoutesLastRoutes).To(ConsistOf(tcpRoutes[0], tcpRoutes[1]))
			Expect(tcpRouter.CreateRoutesLastRouterGroup).To(Equal(tcpRouter.RouterGroupsResult[0]))
		})
	})
})
