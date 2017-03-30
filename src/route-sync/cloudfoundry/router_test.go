package cloudfoundry_test

import (
	"errors"
	. "route-sync/cloudfoundry"
	"route-sync/cloudfoundry/tcp"
	tcpfakes "route-sync/cloudfoundry/tcp/fakes"
	"route-sync/route"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/route-registrar/config"
	messagebusfakes "code.cloudfoundry.org/route-registrar/messagebus/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Router", func() {
	Context("with a valid message bus", func() {
		var (
			router     route.Router
			tcpRouter  tcpfakes.FakeRouter
			messageBus messagebusfakes.FakeMessageBus
			httpRoutes []*route.HTTP
			tcpRoutes  []*route.TCP
			logger     lager.Logger
		)
		BeforeEach(func() {
			logger = lagertest.NewTestLogger("")
			messageBus = messagebusfakes.FakeMessageBus{}
			tcpRouter = tcpfakes.FakeRouter{}
			tcpRouter.RouterGroupsResult = []tcp.RouterGroup{tcp.RouterGroup{
				Guid:            "abc123",
				Name:            "myRouter",
				ReservablePorts: "1000-2000",
				Type:            "tcp",
			},
			}
			router = NewRouter(&messageBus, &tcpRouter)
			httpRoutes = []*route.HTTP{
				&route.HTTP{
					Backends: []route.Endpoint{route.Endpoint{IP: "10.10.10.10", Port: 9090}},
					Name:     "foobar.cf-app.com",
				},
				&route.HTTP{
					Backends: []route.Endpoint{
						route.Endpoint{IP: "10.10.10.10", Port: 9090},
						route.Endpoint{IP: "10.2.2.2", Port: 8080},
					},
					Name: "baz.cf-app.com",
				},
			}

			tcpRoutes = []*route.TCP{
				&route.TCP{
					Backends: []route.Endpoint{route.Endpoint{IP: "10.10.10.10", Port: 9090}},
					Frontend: 1010,
				},
				&route.TCP{
					Backends: []route.Endpoint{
						route.Endpoint{IP: "10.10.10.10", Port: 9090},
						route.Endpoint{IP: "10.2.2.2", Port: 8080},
					},
					Frontend: 2020,
				},
			}
		})
		It("posts a single L7 route", func() {
			Expect(router.HTTP([]*route.HTTP{httpRoutes[0]})).To(Succeed())

			Expect(messageBus.SendMessageCallCount()).To(Equal(1))
			subject, host, route, privateInstanceId := messageBus.SendMessageArgsForCall(0)

			Expect(subject).To(Equal("router.register"))
			Expect(host).To(Equal("10.10.10.10"))
			Expect(route.Port).To(Equal(9090))
			Expect(route.URIs).To(HaveLen(1))
			Expect(route.URIs[0]).To(Equal("foobar.cf-app.com"))
			Expect(privateInstanceId).NotTo(Equal(""))
		})

		It("posts multiple L7 routes with multiple backends", func() {
			Expect(router.HTTP(httpRoutes)).To(Succeed())
			Expect(messageBus.SendMessageCallCount()).To(Equal(3))
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

		Context("when there are multiple router groups", func() {
			BeforeEach(func() {
				tcpRouter.RouterGroupsResult = []tcp.RouterGroup{
					{
						Guid:            "abc123",
						Name:            "myRouter",
						ReservablePorts: "1000-2000",
						Type:            "tcp",
					},
					{
						Guid:            "abc121",
						Name:            "notmyRouter",
						ReservablePorts: "2000-3000",
						Type:            "tcp",
					},
				}
			})

			It("returns error", func() {
				Expect(router.TCP(tcpRoutes)).NotTo(Succeed())
			})
		})

		Context("when TCP Router fails during RouterGroups", func() {
			BeforeEach(func() {
				tcpRouter.RouterGroupsError = errors.New("fail")
			})

			It("returns error", func() {
				Expect(router.TCP(tcpRoutes)).NotTo(Succeed())
			})
		})

		It("connects to nats", func() {
			router.Connect(nil, logger)
			Expect(messageBus.ConnectCallCount()).To(Equal(1))
		})

		It("panics when failing to connect to nats", func() {
			messageBus.ConnectStub = func([]config.MessageBusServer) error {
				return errors.New("Failed to connect")
			}
			defer func() {
				recover()
				Eventually(logger).Should(gbytes.Say("connecting to NATS"))
			}()
			router.Connect(nil, logger)
		})

		It("Returns messagebus errors", func() {
			messageBus.SendMessageStub = func(string, string, config.Route, string) error {
				return errors.New("BOOM!")
			}

			Expect(router.HTTP(httpRoutes)).To(HaveOccurred())
		})
	})
})
