package application

import (
	"context"
	"route-sync/config"
	"route-sync/pooler/poolerfakes"
	"route-sync/route"
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
		logger     lager.Logger
		poolerFake *poolerfakes.FakePooler
		sourceFake *routefakes.FakeSource
		routerFake *routefakes.FakeRouter
		cfg        = &config.Config{
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
		poolerFake = &poolerfakes.FakePooler{}
		sourceFake = &routefakes.FakeSource{}
		routerFake = &routefakes.FakeRouter{}
	})
	Context("Application", func() {
		It("runs the Application and eventually finishes", func() {
			app := NewApplication(logger, poolerFake, sourceFake, routerFake)
			app.Run(context.Background(), cfg)

			Eventually(logger).Should(gbytes.Say("exiting"))
		})

		It("starts the pooler", func() {
			app := NewApplication(logger, poolerFake, sourceFake, routerFake)
			app.Run(context.Background(), cfg)

			Expect(poolerFake.RunCallCount()).To(Equal(1))

			_, startSource, startRouter := poolerFake.RunArgsForCall(0)
			Expect(startSource).To(BeEquivalentTo(sourceFake))
			Expect(startRouter).To(BeEquivalentTo(routerFake))
		})

		It("starts a connection to nats", func() {
			app := NewApplication(logger, poolerFake, sourceFake, routerFake)
			app.Run(context.Background(), cfg)

			Expect(routerFake.ConnectCallCount()).To(Equal(1))
			messageBusServers, log := routerFake.ConnectArgsForCall(0)

			Expect(messageBusServers).To(HaveLen(1))
			Expect(messageBusServers[0].Host).To(Equal("host"))
			Expect(messageBusServers[0].User).To(Equal("user"))
			Expect(messageBusServers[0].Password).To(Equal("password"))

			Expect(log).To(BeEquivalentTo(logger))
		})

		Context("with a pooler that waits till cancel", func() {
			BeforeEach(func() {
				poolerFake.RunStub = func(ctx context.Context, _ route.Source, _ route.Router) {
					<-ctx.Done()
				}
			})
			It("runs until cancelled", func() {
				ctx, cancelFunc := context.WithCancel(context.Background())

				app := NewApplication(logger, poolerFake, sourceFake, routerFake)

				var appRunning bool
				var isAppRunning = func() bool { return appRunning }
				go func() {
					appRunning = true
					app.Run(ctx, cfg)
					appRunning = false
				}()

				Eventually(isAppRunning).Should(BeTrue())
				cancelFunc()
				Eventually(isAppRunning).Should(BeFalse())
			})
		})
	})
})
