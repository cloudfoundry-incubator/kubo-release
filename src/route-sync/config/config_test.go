package config_test

import (
	"io/ioutil"
	"os"
	. "route-sync/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gbytes"

	cfConfig "code.cloudfoundry.org/route-registrar/config"
	uaaconfig "code.cloudfoundry.org/uaa-go-client/config"
)

var _ = Describe("Config", func() {
	Context("Reading schema from a file that doesn't exist", func() {
		It("returns an error", func() {
			cs, err := NewConfigSchemaFromFile("why-do-you-have-this-file")
			Expect(err).To(HaveOccurred())
			Expect(cs).To(BeNil())
		})
	})
	Context("Reading schema from a file", func() {
		var (
			configFile *os.File
		)

		BeforeEach(func() {
			var err error
			configFile, err = ioutil.TempFile("", "routesync")
			if err != nil {
				panic(err)
			}
		})

		AfterEach(func() {
			configFile.Close()
			os.Remove(configFile.Name())
		})

		Context("with a partially filled in file", func() {
			BeforeEach(func() {
				configFile.Write([]byte(`---
app_domain_name: mydomain
uaa_api_url: yoururl
`))
			})
			It("returns a partially filled ConfigSchema", func() {
				cs, err := NewConfigSchemaFromFile(configFile.Name())
				Expect(err).NotTo(HaveOccurred())
				Expect(cs.CloudFoundryAppDomainName).To(Equal("mydomain"))
				Expect(cs.UaaApiUrl).To(Equal("yoururl"))
			})
		})
		Context("with a poorly formated file", func() {
			BeforeEach(func() {
				configFile.Write([]byte(`---
yaml-error
`))
			})
			It("does not return a config schema", func() {
				cs, err := NewConfigSchemaFromFile(configFile.Name())

				Expect(err).To(HaveOccurred())
				Expect(cs).To(BeNil())
			})
		})
	})

	Context("ToSchema", func() {
		var (
			schema *ConfigSchema
		)
		BeforeEach(func() {
			schema = &ConfigSchema{
				NatsServers: []MessageBusServerSchema{
					{Host: "myhost", User: "user", Password: "pass"},
					{Host: "myhost2", User: "user2", Password: "pass2"},
				},
				RoutingApiUrl:             "routingurl",
				CloudFoundryAppDomainName: "appdomain",
				UaaApiUrl:                 "uaaurl",
				RoutingApiUsername:        "myuser",
				RoutingApiClientSecret:    "mysecret",
				SkipTlsVerification:       false,
				KubeConfigPath:            "somewhere",
			}
		})
		Context("with a valid schema", func() {
			var (
				cfg *Config
				err error
			)
			BeforeEach(func() {
				cfg, err = schema.ToConfig()
			})
			It("creates a config object without error", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(cfg).ToNot(BeNil())
			})
			It("parses to a config object", func() {
				Expect(cfg.NatsServers).To(BeEquivalentTo([]cfConfig.MessageBusServer{
					{Host: "myhost", User: "user", Password: "pass"},
					{Host: "myhost2", User: "user2", Password: "pass2"},
				}))
				Expect(cfg.RoutingApiUrl).To(Equal("routingurl"))
				Expect(cfg.CloudFoundryAppDomainName).To(Equal("appdomain"))
				Expect(cfg.UaaApiUrl).To(Equal("uaaurl"))
				Expect(cfg.RoutingApiUsername).To(Equal("myuser"))
				Expect(cfg.RoutingApiClientSecret).To(Equal("mysecret"))
				Expect(cfg.SkipTlsVerification).To(BeFalse())
				Expect(cfg.KubeConfigPath).To(Equal("somewhere"))
			})
			It("can construct a UAAConfig", func() {
				Expect(cfg.UAAConfig()).To(BeEquivalentTo(&uaaconfig.Config{
					ClientName:       "myuser",
					ClientSecret:     "mysecret",
					UaaEndpoint:      "uaaurl",
					SkipVerification: false,
				}))
			})
		})

		DescribeTable("with a field missing", func(fieldName string, removalFunc func(*ConfigSchema)) {
			removalFunc(schema)
			cfg, err := schema.ToConfig()
			Expect(err).To(HaveOccurred())
			Expect(cfg).To(BeNil())
			buf := gbytes.BufferWithBytes([]byte(err.Error()))
			Expect(buf).To(gbytes.Say(fieldName))
		},
			Entry("nil nats_servers", "nats_servers", func(cs *ConfigSchema) { cs.NatsServers = nil }),
			Entry("empty nats_servers", "nats_servers", func(cs *ConfigSchema) { cs.NatsServers = []MessageBusServerSchema{} }),
			Entry("a nats server with an empty host", `nats_servers\[\].host`, func(cs *ConfigSchema) { cs.NatsServers[0].Host = "" }),
			Entry("a nats server with an empty user", `nats_servers\[\].user`, func(cs *ConfigSchema) { cs.NatsServers[0].User = "" }),
			Entry("a nats server with an empty password", `nats_servers\[\].password`, func(cs *ConfigSchema) { cs.NatsServers[0].Password = "" }),
			Entry("empty app_domain_name", "app_domain_name", func(cs *ConfigSchema) { cs.CloudFoundryAppDomainName = "" }),
			Entry("empty uaa_api_url", "uaa_api_url", func(cs *ConfigSchema) { cs.UaaApiUrl = "" }),
			Entry("empty routing_api_url", "routing_api_url", func(cs *ConfigSchema) { cs.RoutingApiUrl = "" }),
			Entry("empty routing_api_username", "routing_api_username", func(cs *ConfigSchema) { cs.RoutingApiUsername = "" }),
			Entry("empty routing_api_client_secret", "routing_api_client_secret", func(cs *ConfigSchema) { cs.RoutingApiClientSecret = "" }),
			Entry("empty kube_config_path", "kube_config_path", func(cs *ConfigSchema) { cs.KubeConfigPath = "" }),
		)
	})
})
