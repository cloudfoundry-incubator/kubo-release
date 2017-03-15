package config_test

import (
	"os"
	. "route-sync/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	BeforeEach(func() {
		os.Setenv("ROUTESYNC_NATS_SERVERS", "[{\"Host\": \"10.0.1.8:4222\",\"User\": \"nats\", \"Password\": \"natspass\"}]")
		os.Setenv("ROUTESYNC_CLOUD_FOUNDRY_API_URL", "https://api.cf.example.org")
		os.Setenv("ROUTESYNC_CLOUD_FOUNDRY_APP_DOMAIN_NAME", "apps.cf.example.org")
		os.Setenv("ROUTESYNC_UAA_API_URL", "https://uaa.cf.example.org")
		os.Setenv("ROUTESYNC_ROUTING_API_USERNAME", "routeUser")
		os.Setenv("ROUTESYNC_ROUTING_API_CLIENT_SECRET", "aabbcc")
		os.Setenv("ROUTESYNC_SKIP_TLS_VERIFICATION", "true")
		os.Setenv("ROUTESYNC_KUBE_CONFIG_PATH", "~/.config/kube")
	})

	It("returns a valid config frm the enviornment", func() {
		c, err := NewConfig()

		Expect(err).To(BeNil())
		Expect(c.NatsServers).To(HaveLen(1))

		natsServer := c.NatsServers[0]
		Expect(natsServer.Host).To(Equal("10.0.1.8:4222"))
		Expect(natsServer.User).To(Equal("nats"))
		Expect(natsServer.Password).To(Equal("natspass"))
	})
})
