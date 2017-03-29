package application

import (
	"route-sync/config"
	"route-sync/pooler/poolerfakes"
	"route-sync/route/routefakes"
	cfConfig "code.cloudfoundry.org/route-registrar/config"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Application", func() {
	var (
		logger lager.Logger
		pooler *poolerfakes.FakePooler
		source *routefakes.FakeSource
		router *routefakes.FakeRouter
		cfg    = &config.Config{
			RawNatsServers:            "[{\"Host\": \"host\",\"User\": \"user\", \"Password\": \"password\"}]",
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
		pooler = &poolerfakes.FakePooler{}
		source = &routefakes.FakeSource{}
		router = &routefakes.FakeRouter{}
	})
	Context("Application", func() {
		It("runs the Application and eventually finishes", func() {
			app := NewApplication(logger, pooler, source, router)
			app.Run(0, cfg)

			Eventually(logger).Should(gbytes.Say("exiting"))
		})

		It("starts the pooler", func() {
			app := NewApplication(logger, pooler, source, router)
			app.Run(0, cfg)

			Expect(pooler.StartCallCount()).To(Equal(1))

			startSource, startRouter := pooler.StartArgsForCall(0)
			Expect(startSource).To(BeEquivalentTo(source))
			Expect(startRouter).To(BeEquivalentTo(router))
		})

		It("starts a connection to nats", func() {
			app := NewApplication(logger, pooler, source, router)
			app.Run(0, cfg)

			Expect(router.ConnectCallCount()).To(Equal(1))
			messageBusServers, log := router.ConnectArgsForCall(0)

			Expect(messageBusServers).To(HaveLen(1))
			Expect(messageBusServers[0].Host).To(Equal("host"))
			Expect(messageBusServers[0].User).To(Equal("user"))
			Expect(messageBusServers[0].Password).To(Equal("password"))

			Expect(log).To(BeEquivalentTo(logger))
		})
	})
})
