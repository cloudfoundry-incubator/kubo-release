package cloudfoundry

import (
	"errors"
	"route-sync/cloudfoundry/tcp"
	tcpfakes "route-sync/cloudfoundry/tcp/fakes"
	"route-sync/config"

	cfConfig "code.cloudfoundry.org/route-registrar/config"

	"route-sync/route"
	"route-sync/route/routefakes"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/route-registrar/messagebus"
	messagebusfakes "code.cloudfoundry.org/route-registrar/messagebus/fakes"
	uaa "code.cloudfoundry.org/uaa-go-client"
	uaaconfig "code.cloudfoundry.org/uaa-go-client/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("CloudFoundryRouterBuilder", func() {
	var (
		logger                    lager.Logger
		fakeUaaClientBuilderFunc  func(logger lager.Logger, cfg *uaaconfig.Config, clock clock.Clock) (uaa.Client, error)
		fakeTcpRouterBuilderFunc  func(uaaClient uaa.Client, routingApiUrl string, skipTlsVerification bool) (tcp.Router, error)
		fakeMessageBusBuilderFunc func(logger lager.Logger) messagebus.MessageBus
		fakeNewRouterFunc         func(bus messagebus.MessageBus, tcpRouter tcp.Router) route.Router
		fakeRouter                route.Router
		fakeTcpRouter             tcp.Router
		fakeMessageBus            *messagebusfakes.FakeMessageBus
		cfg                       = &config.Config{
			NatsServers:               []cfConfig.MessageBusServer{{Host: "host", User: "user", Password: "password"}},
			RoutingApiUrl:             "https://api.cf.example.org",
			CloudFoundryAppDomainName: "apps.cf.example.org",
			UAAApiURL:                 "https://uaa.cf.example.org",
			RoutingAPIUsername:        "routeUser",
			RoutingAPIClientSecret:    "aabbcc",
			SkipTLSVerification:       true,
			KubeConfigPath:            "~/.config/kube",
		}
	)
	BeforeEach(func() {
		logger = lagertest.NewTestLogger("")
		fakeRouter = &routefakes.FakeRouter{}
		fakeTcpRouter = &tcpfakes.FakeRouter{}
		fakeNewRouterFunc = func(bus messagebus.MessageBus, tcpRouter tcp.Router) route.Router {
			return fakeRouter
		}
		fakeUaaClientBuilderFunc = func(logger lager.Logger, cfg *uaaconfig.Config, clock clock.Clock) (uaa.Client, error) {
			return nil, nil
		}
		fakeTcpRouterBuilderFunc = func(uaaClient uaa.Client, routingApiUrl string, skipTlsVerification bool) (tcp.Router, error) {
			return fakeTcpRouter, nil
		}
		fakeMessageBus = &messagebusfakes.FakeMessageBus{}
		fakeMessageBusBuilderFunc = func(logger lager.Logger) messagebus.MessageBus {
			return fakeMessageBus
		}
	})

	Context("TCP Router", func() {
		It("returns a TCP router", func() {
			var usedRouter tcp.Router
			fakeNewRouterFunc = func(_ messagebus.MessageBus, tcpRouter tcp.Router) route.Router {
				usedRouter = tcpRouter
				return nil
			}
			routingBuilder := NewRouterBuilder(fakeUaaClientBuilderFunc, fakeTcpRouterBuilderFunc, fakeMessageBusBuilderFunc, fakeNewRouterFunc)
			routingBuilder.CreateRouter(cfg, logger)
			Expect(usedRouter).To(Equal(fakeTcpRouter))
		})

		It("panics when uaa client fails", func() {
			fakeUaaClientBuilderFunc = func(logger lager.Logger, cfg *uaaconfig.Config, clock clock.Clock) (uaa.Client, error) {
				return nil, errors.New("")
			}
			routingBuilder := NewRouterBuilder(fakeUaaClientBuilderFunc, fakeTcpRouterBuilderFunc, fakeMessageBusBuilderFunc, fakeNewRouterFunc)
			var createTcpRouter = func() { routingBuilder.CreateRouter(cfg, logger) }

			Expect(createTcpRouter).To(Panic())
			Expect(logger).To(gbytes.Say("creating UAA client"))
		})
		It("panics when tcp router creation fails", func() {
			fakeTcpRouterBuilderFunc = func(uaaClient uaa.Client, routingApiUrl string, skipTlsVerification bool) (tcp.Router, error) {
				return nil, errors.New("")
			}
			routingBuilder := NewRouterBuilder(fakeUaaClientBuilderFunc, fakeTcpRouterBuilderFunc, fakeMessageBusBuilderFunc, fakeNewRouterFunc)
			var createTcpRouter = func() { routingBuilder.CreateRouter(cfg, logger) }

			Expect(createTcpRouter).To(Panic())
			Expect(logger).To(gbytes.Say("creating TCP router"))
		})
	})

	Context("HTTP Router", func() {
		It("embeds the HTTP router", func() {
			var messageBus messagebus.MessageBus
			fakeNewRouterFunc = func(bus messagebus.MessageBus, _ tcp.Router) route.Router {
				messageBus = bus
				return fakeRouter
			}
			routingBuilder := NewRouterBuilder(fakeUaaClientBuilderFunc, fakeTcpRouterBuilderFunc, fakeMessageBusBuilderFunc, fakeNewRouterFunc)
			routingBuilder.CreateRouter(cfg, logger)

			Expect(messageBus).To(Equal(fakeMessageBus))
		})
	})

	It("returns a combined router", func() {
		routingBuilder := NewRouterBuilder(fakeUaaClientBuilderFunc, fakeTcpRouterBuilderFunc, fakeMessageBusBuilderFunc, fakeNewRouterFunc)
		router := routingBuilder.CreateRouter(cfg, logger)

		Expect(router).To(Equal(fakeRouter))
	})
})
