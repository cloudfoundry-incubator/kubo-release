package config_test

import (
	"io/ioutil"
	"os"
	. "route-sync/config"

	cfConfig "code.cloudfoundry.org/route-registrar/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
				RoutingApiUrl:             "routingurl",
				CloudFoundryAppDomainName: "appdomain",
				UAAApiURL:                 "uaaurl",
				RoutingAPIUsername:        "myuser",
				RoutingAPIClientSecret:    "mysecret",
				SkipTLSVerification:       false,
				KubeConfigPath:            "somewhere",
			}
		})
		Context("with a valid schema", func() {
			It("parses to a config object", func() {
				cfg, err := schema.ToConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(cfg).ToNot(BeNil())
				Expect(cfg.NatsServers).To(BeEquivalentTo([]cfConfig.MessageBusServer{
					{Host: "myhost", User: "user", Password: "pass"},
					{Host: "myhost2", User: "user2", Password: "pass2"},
				}))
				Expect(cfg.RoutingApiUrl).To(Equal("routingurl"))
				Expect(cfg.CloudFoundryAppDomainName).To(Equal("appdomain"))
				Expect(cfg.UAAApiURL).To(Equal("uaaurl"))
				Expect(cfg.RoutingAPIUsername).To(Equal("myuser"))
				Expect(cfg.RoutingAPIClientSecret).To(Equal("mysecret"))
				Expect(cfg.SkipTLSVerification).To(BeFalse())
				Expect(cfg.KubeConfigPath).To(Equal("somewhere"))
			})
		})
	})
})
