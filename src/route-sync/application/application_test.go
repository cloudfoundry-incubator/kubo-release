package application

import (
	"route-sync/config"
	"route-sync/pooler/poolerfakes"
	"route-sync/route/routefakes"

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
		src    *routefakes.FakeSource
		router *routefakes.FakeRouter
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
		pooler = &poolerfakes.FakePooler{}
		src = &routefakes.FakeSource{}
		router = &routefakes.FakeRouter{}
	})
	Context("Application", func() {
		It("runs the Application and eventually finishes", func() {
			app := NewApplication(logger, pooler, src, router)
			app.Run(0, cfg)

			Eventually(logger).Should(gbytes.Say("exiting"))
		})
		It("starts the pooler", func() {
			app := NewApplication(logger, pooler, src, router)
			app.Run(0, cfg)

			Expect(pooler.StartCallCount()).To(Equal(1))
		})
		It("starts a connection to nats", func() {
			app := NewApplication(logger, pooler, src, router)
			app.Run(0, cfg)

			Expect(router.ConnectCallCount()).To(Equal(1))
		})
	})
})
