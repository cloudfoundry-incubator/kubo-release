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
