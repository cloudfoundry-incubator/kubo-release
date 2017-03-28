package cloudfoundry

import (
	"errors"
	"route-sync/cloudfoundry/tcp"
	tcpfakes "route-sync/cloudfoundry/tcp/fakes"
	"route-sync/config"

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
		logger lager.Logger
		cfg    = &config.Config{
			RawNatsServers:            "[{\"Host\": \"10.0.1.8:4222\",\"User\": \"nats\", \"Password\": \"natspass\"}]",
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
	})
	Context("TCPRouter", func() {
		var (
			fakeNewUAAClient func(logger lager.Logger, cfg *uaaconfig.Config, clock clock.Clock) (uaa.Client, error)
			fakeNewTCPRouter func(uaaClient uaa.Client, routingApiUrl string, skipTlsVerification bool) (tcp.Router, error)
			fakeRouter       tcp.Router
		)
		BeforeEach(func() {
			fakeRouter = &tcpfakes.FakeRouter{}
			fakeNewUAAClient = func(logger lager.Logger, cfg *uaaconfig.Config, clock clock.Clock) (uaa.Client, error) {
				return nil, nil
			}
			fakeNewTCPRouter = func(uaaClient uaa.Client, routingApiUrl string, skipTlsVerification bool) (tcp.Router, error) {
				return fakeRouter, nil
			}
		})
		It("returns a TCP router", func() {
			router := NewCloudFoundryRoutingBuilder(cfg, logger)
			client := router.GetTCPRouter(fakeNewUAAClient, fakeNewTCPRouter)
			Expect(client).To(Equal(fakeRouter))
		})
		It("panics when uaa client fails", func() {
			fakeNewUAAClient = func(logger lager.Logger, cfg *uaaconfig.Config, clock clock.Clock) (uaa.Client, error) {
				return nil, errors.New("")
			}
			router := NewCloudFoundryRoutingBuilder(cfg, logger)
			defer func() {
				recover()
				Eventually(logger).Should(gbytes.Say("creating UAA client"))
			}()
			router.GetTCPRouter(fakeNewUAAClient, fakeNewTCPRouter)
		})
		It("panics when tcp router creation fails", func() {
			fakeNewTCPRouter = func(uaaClient uaa.Client, routingApiUrl string, skipTlsVerification bool) (tcp.Router, error) {
				return nil, errors.New("")
			}
			router := NewCloudFoundryRoutingBuilder(cfg, logger)
			defer func() {
				recover()
				Eventually(logger).Should(gbytes.Say("creating TCP router"))
			}()
			router.GetTCPRouter(fakeNewUAAClient, fakeNewTCPRouter)
		})
	})

	Context("HTTPRouter", func() {
		var (
			fakeMessageBus    *messagebusfakes.FakeMessageBus
			fakeNewMessageBus func(logger lager.Logger) messagebus.MessageBus
		)
		BeforeEach(func() {
			fakeMessageBus = &messagebusfakes.FakeMessageBus{}
			fakeNewMessageBus = func(logger lager.Logger) messagebus.MessageBus {
				return fakeMessageBus
			}
		})
		It("returns correct HTTP router", func() {
			router := NewCloudFoundryRoutingBuilder(cfg, logger)
			mb := router.GetHTTPRouter(fakeNewMessageBus)
			Expect(mb).To(Equal(fakeMessageBus))
		})
	})
})
