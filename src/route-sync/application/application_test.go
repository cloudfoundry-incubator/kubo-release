package application

import (
	"encoding/json"
	"route-sync/config"

	"code.cloudfoundry.org/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application", func() {
	var (
		cfg = &config.Config{
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
		json.Unmarshal([]byte(cfg.RawNatsServers), &cfg.NatsServers)
	})
	Context("CreateSink", func() {
		It("Should create a TCP router from a config", func() {
			app := NewApp()
			Expect(true).To(BeTrue())
			logger := lagertest.NewTestLogger("app-test")
			router, err := app.NewTCPRouter(logger, cfg)
			Expect(router).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})
		It("Should fail to create a TCP router from an invalid config", func() {
			Expect(true).To(BeTrue())
			app := NewApp()
			logger := lagertest.NewTestLogger("app-test")
			router, err := app.NewTCPRouter(logger, cfg)
			Expect(router).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})
})
