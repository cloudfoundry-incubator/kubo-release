package config_test

import (
	"fmt"
	"io/ioutil"
	"os"
	. "route-sync/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	cfConfig "code.cloudfoundry.org/route-registrar/config"
	uaaconfig "code.cloudfoundry.org/uaa-go-client/config"
)

var _ = Describe("Config", func() {
	Context("Reading schema from a file that doesn't exist", func() {
		It("does not return a config schema", func() {
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

		Context("with an empty file", func() {
			It("returns a ConfigSchema", func() {
				cs, err := NewConfigSchemaFromFile(configFile.Name())
				Expect(err).NotTo(HaveOccurred())
				Expect(cs).NotTo(BeNil())
			})
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
				Expect(cs.UAAApiURL).To(Equal("yoururl"))
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
				RoutingAPIURL:             "routingurl",
				CloudFoundryAppDomainName: "appdomain",
				UAAApiURL:                 "uaaurl",
				RoutingAPIUsername:        "myuser",
				RoutingAPIClientSecret:    "mysecret",
				SkipTLSVerification:       false,
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
				Expect(cfg.RoutingAPIURL).To(Equal("routingurl"))
				Expect(cfg.CloudFoundryAppDomainName).To(Equal("appdomain"))
				Expect(cfg.UAAApiURL).To(Equal("uaaurl"))
				Expect(cfg.RoutingAPIUsername).To(Equal("myuser"))
				Expect(cfg.RoutingAPIClientSecret).To(Equal("mysecret"))
				Expect(cfg.SkipTLSVerification).To(BeFalse())
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

		fieldRequiredTests := []struct {
			fieldName   string
			removalFunc func(*ConfigSchema)
		}{
			{"nats_servers", func(cs *ConfigSchema) { cs.NatsServers = nil }},
			{"nats_servers", func(cs *ConfigSchema) { cs.NatsServers = []MessageBusServerSchema{} }},
			{`nats_servers\[\].host`, func(cs *ConfigSchema) { cs.NatsServers[0].Host = "" }},
			{`nats_servers\[\].user`, func(cs *ConfigSchema) { cs.NatsServers[0].User = "" }},
			{`nats_servers\[\].password`, func(cs *ConfigSchema) { cs.NatsServers[0].Password = "" }},
			{"app_domain_name", func(cs *ConfigSchema) { cs.CloudFoundryAppDomainName = "" }},
			{"uaa_api_url", func(cs *ConfigSchema) { cs.UAAApiURL = "" }},
			{"routing_api_url", func(cs *ConfigSchema) { cs.RoutingAPIURL = "" }},
			{"routing_api_username", func(cs *ConfigSchema) { cs.RoutingAPIUsername = "" }},
			{"routing_api_client_secret", func(cs *ConfigSchema) { cs.RoutingAPIClientSecret = "" }},
			{"kube_config_path", func(cs *ConfigSchema) { cs.KubeConfigPath = "" }},
		}

		for _, testCase := range fieldRequiredTests {
			// Create a fresh copy of testCase so each func points to a unique value
			testCase := testCase
			Context(fmt.Sprintf("With field %s missing", testCase.fieldName), func() {
				It("returns an error", func() {
					testCase.removalFunc(schema)
					cfg, err := schema.ToConfig()
					Expect(err).To(HaveOccurred())
					Expect(cfg).To(BeNil())
					buf := gbytes.BufferWithBytes([]byte(err.Error()))
					Expect(buf).To(gbytes.Say(testCase.fieldName))
				})
			})
		}
	})
})
