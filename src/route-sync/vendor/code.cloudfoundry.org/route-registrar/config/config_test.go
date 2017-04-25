package config_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"code.cloudfoundry.org/route-registrar/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Config", func() {
	var (
		configSchema config.ConfigSchema

		registrationInterval0String string
		registrationInterval1String string

		registrationInterval0 time.Duration
		registrationInterval1 time.Duration

		routeName0 string
		routeName1 string

		port0 int
		port1 int
	)

	BeforeEach(func() {
		registrationInterval0String = "20s"
		registrationInterval1String = "10s"

		registrationInterval0 = 20 * time.Second
		registrationInterval1 = 10 * time.Second

		routeName0 = "route-0"
		routeName1 = "route-1"

		port0 = 3000
		port1 = 3001

		configSchema = config.ConfigSchema{
			MessageBusServers: []config.MessageBusServerSchema{
				config.MessageBusServerSchema{
					Host:     "some-host",
					User:     "some-user",
					Password: "some-password",
				},
				config.MessageBusServerSchema{
					Host:     "another-host",
					User:     "another-user",
					Password: "another-password",
				},
			},
			Routes: []config.RouteSchema{
				{
					Name:                 routeName0,
					Port:                 &port0,
					RegistrationInterval: registrationInterval0String,
					URIs:                 []string{"my-app.my-domain.com"},
				},
				{
					Name:                 routeName1,
					Port:                 &port1,
					RegistrationInterval: registrationInterval1String,
					URIs:                 []string{"my-other-app.my-domain.com"},
				},
			},
			Host: "127.0.0.1",
		}
	})

	Describe("NewConfigSchemaFromFile", func() {
		It("returns a valid config", func() {
			cfg_file := "../example_config/example.yml"
			cfg, err := config.NewConfigSchemaFromFile(cfg_file)

			Expect(err).NotTo(HaveOccurred())
			Expect(cfg).To(Equal(configSchema))
		})

		Context("when the file does not exists", func() {
			It("returns an error", func() {
				cfg_file := "notexist"
				_, err := config.NewConfigSchemaFromFile(cfg_file)

				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the config is invalid", func() {
			var (
				configFile *os.File
			)

			BeforeEach(func() {
				var err error
				configFile, err = ioutil.TempFile("", "route-registrar-config")
				Expect(err).NotTo(HaveOccurred())

				_, err = configFile.Write([]byte("invalid yaml %^&#"))
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				os.Remove(configFile.Name())
			})

			It("returns an error", func() {
				_, err := config.NewConfigSchemaFromFile(configFile.Name())
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("ToConfig", func() {
		It("returns a Config object and no error", func() {
			c, err := configSchema.ToConfig()
			Expect(err).ToNot(HaveOccurred())

			expectedC := &config.Config{
				Host: configSchema.Host,
				MessageBusServers: []config.MessageBusServer{
					{
						Host:     configSchema.MessageBusServers[0].Host,
						User:     configSchema.MessageBusServers[0].User,
						Password: configSchema.MessageBusServers[0].Password,
					},
					{
						Host:     configSchema.MessageBusServers[1].Host,
						User:     configSchema.MessageBusServers[1].User,
						Password: configSchema.MessageBusServers[1].Password,
					},
				},
				Routes: []config.Route{
					{
						Name:                 routeName0,
						Port:                 port0,
						RegistrationInterval: registrationInterval0,
						URIs:                 configSchema.Routes[0].URIs,
					},
					{
						Name:                 routeName1,
						Port:                 port1,
						RegistrationInterval: registrationInterval1,
						URIs:                 configSchema.Routes[1].URIs,
					},
				},
			}

			Expect(c).To(Equal(expectedC))
		})

		Describe("Routes", func() {
			Context("when config input includes route_service_url", func() {
				BeforeEach(func() {
					configSchema.Routes[0].RouteServiceUrl = "https://rs.example.com"
					configSchema.Routes[1].RouteServiceUrl = "https://rs.example.com"
				})

				It("includes them in the config", func() {
					c, err := configSchema.ToConfig()
					Expect(err).ToNot(HaveOccurred())

					Expect(c.Routes[0].RouteServiceUrl).Should(Equal(configSchema.Routes[0].RouteServiceUrl))
					Expect(c.Routes[1].RouteServiceUrl).Should(Equal(configSchema.Routes[1].RouteServiceUrl))
				})

				Context("and the route_service_url is not a valid URI", func() {
					BeforeEach(func() {
						configSchema.Routes[0].RouteServiceUrl = "ht%tp://rs.example.com"
					})

					It("returns an error", func() {
						_, err := configSchema.ToConfig()
						Expect(err).To(HaveOccurred())
					})
				})
			})

			Context("when the config input includes tags", func() {
				BeforeEach(func() {
					configSchema.Routes[0].Tags = map[string]string{"key": "value"}
					configSchema.Routes[1].Tags = map[string]string{"key": "value2"}
				})

				It("includes them in the config", func() {
					c, err := configSchema.ToConfig()
					Expect(err).ToNot(HaveOccurred())

					Expect(c.Routes[0].Tags).Should(Equal(configSchema.Routes[0].Tags))
					Expect(c.Routes[1].Tags).Should(Equal(configSchema.Routes[1].Tags))
				})
			})

			Context("healthcheck is provided", func() {
				BeforeEach(func() {
					configSchema.Routes[0].HealthCheck = &config.HealthCheckSchema{
						Name:       "my healthcheck",
						ScriptPath: "/some/script/path",
					}
				})

				It("sets the healthcheck for the route", func() {
					c, err := configSchema.ToConfig()
					Expect(err).NotTo(HaveOccurred())
					Expect(c.Routes[0].HealthCheck.Name).To(Equal("my healthcheck"))
					Expect(c.Routes[0].HealthCheck.ScriptPath).To(Equal("/some/script/path"))
				})

				Context("The healthcheck timeout is empty", func() {
					BeforeEach(func() {
						configSchema.Routes[0].HealthCheck.Timeout = ""
					})

					It("defaults the healthcheck timeout to half the registration interval", func() {
						c, err := configSchema.ToConfig()
						Expect(err).NotTo(HaveOccurred())

						Expect(c.Routes[0].HealthCheck.Timeout).To(Equal(registrationInterval0 / 2))
					})
				})

				Context("and the healthcheck timeout is provided", func() {
					BeforeEach(func() {
						configSchema.Routes[0].HealthCheck.Timeout = "11s"
					})

					It("sets the healthcheck timeout on the config", func() {
						c, err := configSchema.ToConfig()
						Expect(err).NotTo(HaveOccurred())

						Expect(err).NotTo(HaveOccurred())
						Expect(c.Routes[0].HealthCheck.Timeout).To(Equal(11 * time.Second))
					})
				})
			})
		})
	})

	Describe("Handling errors", func() {
		Describe("on the registration interval", func() {
			Context("The registration interval is empty", func() {
				BeforeEach(func() {
					configSchema.Routes[0].RegistrationInterval = ""
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					buf := gbytes.BufferWithBytes([]byte(err.Error()))
					Expect(buf).To(gbytes.Say(`error with 'route "route-0"'`))
					Expect(buf).To(gbytes.Say("no registration_interval"))
				})
			})

			Context("The registration interval is zero", func() {
				BeforeEach(func() {
					configSchema.Routes[0].RegistrationInterval = "0s"
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					buf := gbytes.BufferWithBytes([]byte(err.Error()))
					Expect(buf).To(gbytes.Say(`error with 'route "route-0"'`))
					Expect(buf).To(gbytes.Say("invalid registration_interval: interval must be greater than 0"))
				})
			})

			Context("The registration interval is negative", func() {
				BeforeEach(func() {
					configSchema.Routes[0].RegistrationInterval = "-1s"
				})

				It("returns an error", func() {
					_, err := configSchema.ToConfig()
					Expect(err).To(HaveOccurred())
					buf := gbytes.BufferWithBytes([]byte(err.Error()))
					Expect(buf).To(gbytes.Say(`error with 'route "route-0"'`))
					Expect(buf).To(gbytes.Say("invalid registration_interval: interval must be greater than 0"))
				})
			})

			Context("When the registration interval has no units", func() {
				BeforeEach(func() {
					configSchema.Routes[0].RegistrationInterval = "1"
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`route "route-0"`))
					Expect(err.Error()).To(ContainSubstring("registration_interval: time: missing unit in duration 1"))
				})
			})

			Context("When the registration interval is not parsable", func() {
				BeforeEach(func() {
					configSchema.Routes[0].RegistrationInterval = "asdf"
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`route "route-0"`))
					Expect(err.Error()).To(ContainSubstring("invalid registration_interval: time: invalid duration asdf"))
				})
			})
		})

		Describe("on route names", func() {
			Context("when the config input does not include a name", func() {
				BeforeEach(func() {
					configSchema.Routes[0].Name = ""
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("route 0"))
					Expect(err.Error()).To(ContainSubstring("no name"))
				})
			})
		})

		Describe("on route ports", func() {
			Context("when the port value is not provided", func() {
				BeforeEach(func() {
					configSchema.Routes[0].Port = nil
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`route "route-0"`))
					Expect(err.Error()).To(ContainSubstring("no port"))
				})
			})

			Context("when the port value is 0", func() {
				BeforeEach(func() {
					zero := 0
					configSchema.Routes[0].Port = &zero
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`route "route-0"`))
					Expect(err.Error()).To(ContainSubstring("invalid port: 0"))
				})
			})

			Context("when the port value is negative", func() {
				BeforeEach(func() {
					negativeValue := -1
					configSchema.Routes[0].Port = &negativeValue
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`route "route-0"`))
					Expect(err.Error()).To(ContainSubstring("invalid port: -1"))
				})
			})
		})

		Describe("on route URIs", func() {
			Context("when the URIs are not provided", func() {
				BeforeEach(func() {
					configSchema.Routes[0].URIs = nil
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`error with 'route "route-0"'`))
					Expect(err.Error()).To(ContainSubstring("* no URIs"))
				})
			})

			Context("when the URIs are empty", func() {
				BeforeEach(func() {
					configSchema.Routes[0].URIs = []string{}
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`error with 'route "route-0"'`))
					Expect(err.Error()).To(ContainSubstring("* no URIs"))
				})
			})

			Context("when the URIs contain empty strings", func() {
				BeforeEach(func() {
					configSchema.Routes[0].URIs = []string{"", "valid-uri", ""}
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`error with 'route "route-0"'`))
					Expect(err.Error()).To(ContainSubstring("* empty URIs"))
				})
			})
		})

		Describe("on the healthcheck, assuming healthcheck is provided", func() {
			BeforeEach(func() {
				configSchema.Routes[0].HealthCheck = &config.HealthCheckSchema{
					Name:       "my healthcheck",
					ScriptPath: "/some/script/path",
				}
			})

			Context("when there is an error with either name or script path and timeout is valid", func() {
				BeforeEach(func() {
					configSchema.Routes[0].HealthCheck.Name = ""
					configSchema.Routes[0].HealthCheck.Timeout = "1s"
				})

				It("still returns the error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`error with 'route "route-0"'`))
					Expect(err.Error()).To(ContainSubstring("error with 'healthcheck'"))
					Expect(err.Error()).To(ContainSubstring("* no name"))
				})
			})

			Context("when the name is empty", func() {
				BeforeEach(func() {
					configSchema.Routes[0].HealthCheck.Name = ""
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`error with 'route "route-0"'`))
					Expect(err.Error()).To(ContainSubstring("error with 'healthcheck'"))
					Expect(err.Error()).To(ContainSubstring("* no name"))
				})
			})

			Context("when the script path is empty", func() {
				BeforeEach(func() {
					configSchema.Routes[0].HealthCheck.ScriptPath = ""
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`error with 'route "route-0"'`))
					Expect(err.Error()).To(ContainSubstring("error with 'healthcheck'"))
					Expect(err.Error()).To(ContainSubstring("* no script_path"))
				})
			})

			Context("when the healthcheck has multiple errors", func() {
				BeforeEach(func() {
					configSchema.Routes[0].HealthCheck.Name = ""
					configSchema.Routes[0].HealthCheck.ScriptPath = ""
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`2 errors with 'route "route-0"'`))
					Expect(err.Error()).To(ContainSubstring("2 errors with 'healthcheck'"))
					Expect(err.Error()).To(ContainSubstring("* no name"))
					Expect(err.Error()).To(ContainSubstring("* no script_path"))
				})
			})

			Context("The healthcheck timeout is zero", func() {
				BeforeEach(func() {
					configSchema.Routes[0].HealthCheck.Timeout = "0s"
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`error with 'route "route-0"'`))
					Expect(err.Error()).To(ContainSubstring("error with 'healthcheck'"))
					Expect(err.Error()).To(ContainSubstring("invalid healthcheck timeout"))
				})
			})

			Context("The healthcheck timeout is negative", func() {
				BeforeEach(func() {
					configSchema.Routes[0].HealthCheck.Timeout = "-1s"
				})

				It("returns an error", func() {
					_, err := configSchema.ToConfig()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`error with 'route "route-0"'`))
					Expect(err.Error()).To(ContainSubstring("error with 'healthcheck'"))
					Expect(err.Error()).To(ContainSubstring("invalid healthcheck timeout: -1s"))
				})
			})

			Context("When the healthcheck timeout has no units", func() {
				BeforeEach(func() {
					configSchema.Routes[0].HealthCheck.Timeout = "1"
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`error with 'route "route-0"'`))
					Expect(err.Error()).To(ContainSubstring("error with 'healthcheck'"))
					Expect(err.Error()).To(ContainSubstring("invalid healthcheck timeout: time: missing unit"))
				})
			})

			Context("When the healthcheck is equal to the registration interval", func() {
				BeforeEach(func() {
					configSchema.Routes[0].HealthCheck.Timeout = configSchema.Routes[0].RegistrationInterval
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(err).To(HaveOccurred())
					Expect(c).To(BeNil())
					Expect(err.Error()).To(ContainSubstring(`error with 'route "route-0"'`))
					Expect(err.Error()).To(ContainSubstring("error with 'healthcheck'"))
					Expect(err.Error()).To(ContainSubstring(
						fmt.Sprintf(
							"invalid healthcheck timeout: %s must be less than the registration interval: %s",
							configSchema.Routes[0].HealthCheck.Timeout,
							configSchema.Routes[0].RegistrationInterval,
						),
					))
				})
			})

			Context("When the healthcheck is greater than the registration interval", func() {
				BeforeEach(func() {
					configSchema.Routes[0].HealthCheck.Timeout = "59s"
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(err).To(HaveOccurred())
					Expect(c).To(BeNil())
					Expect(err.Error()).To(ContainSubstring(`error with 'route "route-0"'`))
					Expect(err.Error()).To(ContainSubstring("error with 'healthcheck'"))
					Expect(err.Error()).To(ContainSubstring(
						fmt.Sprintf(
							"invalid healthcheck timeout: %s must be less than the registration interval: %s",
							configSchema.Routes[0].HealthCheck.Timeout,
							configSchema.Routes[0].RegistrationInterval,
						),
					))
				})
			})

			Context("When the healthcheck timeout is not parsable", func() {
				BeforeEach(func() {
					configSchema.Routes[0].HealthCheck.Timeout = "asdf"
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(`error with 'route "route-0"'`))
					Expect(err.Error()).To(ContainSubstring("error with 'healthcheck'"))
					Expect(err.Error()).To(ContainSubstring("invalid healthcheck timeout: time: invalid duration"))
				})
			})
		})

		Describe("on the host", func() {
			Context("when the host is empty", func() {
				BeforeEach(func() {
					configSchema.Host = ""
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("host required"))
				})
			})
		})

		Describe("on the message bus servers", func() {
			Context("when message bus servers are empty", func() {
				BeforeEach(func() {
					configSchema.MessageBusServers = []config.MessageBusServerSchema{}
				})

				It("returns an error", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("message_bus_servers must have at least one entry"))
				})
			})
		})

		Context("when there are many errors", func() {
			BeforeEach(func() {
				configSchema.Host = ""
				configSchema.MessageBusServers = []config.MessageBusServerSchema{}
				configSchema.Routes[0].RegistrationInterval = ""
				configSchema.Routes[0].Name = ""
				configSchema.Routes[1].RegistrationInterval = ""
				configSchema.Routes[1].Name = ""
			})

			It("displays a count of the errors", func() {
				c, err := configSchema.ToConfig()
				Expect(c).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("there were 6 errors with 'config'"))
			})

			It("aggregates the errors", func() {
				c, err := configSchema.ToConfig()
				Expect(c).To(BeNil())
				Expect(err).To(HaveOccurred())
				buf := gbytes.BufferWithBytes([]byte(err.Error()))
				Expect(buf).To(gbytes.Say("host required"))
				Expect(buf).To(gbytes.Say("message_bus_servers must have at least one entry"))
				Expect(buf).To(gbytes.Say(`there were 2 errors with 'route 0'`))
				Expect(buf).To(gbytes.Say("no name"))
				Expect(buf).To(gbytes.Say("no registration_interval"))
				Expect(buf).To(gbytes.Say(`there were 2 errors with 'route 1'`))
				Expect(buf).To(gbytes.Say("no name"))
				Expect(buf).To(gbytes.Say("no registration_interval"))
			})

			Context("when a registration interval is unparseable", func() {
				BeforeEach(func() {
					configSchema.Routes[1].RegistrationInterval = "asdf"
				})

				It("only returns one error for a given registration_interval", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					buf := gbytes.BufferWithBytes([]byte(err.Error()))
					Expect(buf).To(gbytes.Say(`there were 2 errors with 'route 0'`))
					Expect(buf).To(gbytes.Say(`there were 2 errors with 'route 1'`))
					Expect(buf).To(gbytes.Say("invalid registration_interval: time: invalid duration"))
				})
			})

			Context("when the registration interval does not have units", func() {
				BeforeEach(func() {
					configSchema.Routes[1].RegistrationInterval = "10"
				})

				It("only returns one error for a given registration_interval", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					buf := gbytes.BufferWithBytes([]byte(err.Error()))
					Expect(buf).To(gbytes.Say(`there were 2 errors with 'route 0'`))
					Expect(buf).To(gbytes.Say(`there were 2 errors with 'route 1'`))
					Expect(buf).To(gbytes.Say("invalid registration_interval: time: missing unit in duration 10"))
				})
			})

			Context("when the registration interval has a negative value", func() {
				BeforeEach(func() {
					configSchema.Routes[1].RegistrationInterval = "-10s"
				})

				It("only returns one error for a given registration_interval", func() {
					c, err := configSchema.ToConfig()
					Expect(c).To(BeNil())
					Expect(err).To(HaveOccurred())
					buf := gbytes.BufferWithBytes([]byte(err.Error()))
					Expect(buf).To(gbytes.Say(`there were 2 errors with 'route 1'`))
					Expect(buf).To(gbytes.Say("invalid registration_interval: interval must be greater than 0"))
				})
			})

			Context("and there is a healthcheck", func() {
				BeforeEach(func() {
					configSchema.Routes[1].HealthCheck = &config.HealthCheckSchema{
						Name:       "healthcheck name",
						ScriptPath: "/path/to/script",
					}
				})

				Context("when the registration interval is present", func() {
					BeforeEach(func() {
						configSchema.Routes[1].RegistrationInterval = "10s"
					})

					Context("and the healthcheck is greater than the registration interval", func() {
						var timeoutString = "20s"

						BeforeEach(func() {
							configSchema.Routes[1].HealthCheck.Timeout = timeoutString
						})

						It("returns an error for the timeout", func() {
							c, err := configSchema.ToConfig()
							Expect(c).To(BeNil())
							Expect(err).To(HaveOccurred())
							buf := gbytes.BufferWithBytes([]byte(err.Error()))
							Expect(buf).To(gbytes.Say(`there were 2 errors with 'route 1'`))
							Expect(buf).To(gbytes.Say(fmt.Sprintf(
								"invalid healthcheck timeout: %s must be less than the registration interval",
								timeoutString,
							)))
						})
					})
				})

				Context("and the healthcheck has a zero timeout", func() {
					BeforeEach(func() {
						configSchema.Routes[1].HealthCheck.Timeout = "0s"
					})

					It("returns an error for the timeout", func() {
						c, err := configSchema.ToConfig()
						Expect(c).To(BeNil())
						Expect(err).To(HaveOccurred())
						buf := gbytes.BufferWithBytes([]byte(err.Error()))
						Expect(buf).To(gbytes.Say(`there were 3 errors with 'route 1'`))
						Expect(buf).To(gbytes.Say("invalid healthcheck timeout: 0"))
					})
				})

				Context("and the healthcheck has a negative timeout", func() {
					var timeoutString = "-1s"

					BeforeEach(func() {
						configSchema.Routes[1].HealthCheck.Timeout = timeoutString
					})

					It("returns an error for the timeout", func() {
						c, err := configSchema.ToConfig()
						Expect(c).To(BeNil())
						Expect(err).To(HaveOccurred())
						buf := gbytes.BufferWithBytes([]byte(err.Error()))
						Expect(buf).To(gbytes.Say(`there were 3 errors with 'route 1'`))
						Expect(buf).To(gbytes.Say(fmt.Sprintf(
							"invalid healthcheck timeout: %s",
							timeoutString,
						)))
					})
				})

				Context("and the healthcheck has a valid timeout", func() {
					var timeoutString = "1s"

					BeforeEach(func() {
						configSchema.Routes[1].HealthCheck.Timeout = timeoutString
					})

					It("does not return an error for the timeout", func() {
						c, err := configSchema.ToConfig()
						Expect(c).To(BeNil())
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).ToNot(ContainSubstring(
							"route 'route-1' has invalid healthcheck timeout",
						))
					})
				})

				Context("and the healthcheck has no timeout", func() {
					var timeoutString = ""

					BeforeEach(func() {
						configSchema.Routes[1].HealthCheck.Timeout = timeoutString
					})

					It("returns an error for the timeout", func() {
						c, err := configSchema.ToConfig()
						Expect(c).To(BeNil())
						Expect(err).To(HaveOccurred())
						buf := gbytes.BufferWithBytes([]byte(err.Error()))
						Expect(buf).To(gbytes.Say(`there were 3 errors with 'route 1'`))
						Expect(buf).To(gbytes.Say("invalid healthcheck timeout: time: invalid duration"))
					})
				})

				Context("and the timeout does not parse", func() {
					var timeoutString = "nope"

					BeforeEach(func() {
						configSchema.Routes[1].HealthCheck.Timeout = timeoutString
					})

					It("returns an error for the timeout", func() {
						c, err := configSchema.ToConfig()
						Expect(c).To(BeNil())
						Expect(err).To(HaveOccurred())
						buf := gbytes.BufferWithBytes([]byte(err.Error()))
						Expect(buf).To(gbytes.Say(`there were 3 errors with 'route 1'`))
						Expect(buf).To(gbytes.Say(fmt.Sprintf(
							"invalid healthcheck timeout: time: invalid duration %s",
							timeoutString,
						)))
					})
				})
			})
		})
	})
})
