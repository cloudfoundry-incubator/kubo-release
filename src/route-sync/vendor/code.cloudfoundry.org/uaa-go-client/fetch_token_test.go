package uaa_go_client_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/lager/lagertest"
	trace "code.cloudfoundry.org/trace-logger"
	"code.cloudfoundry.org/uaa-go-client"
	"code.cloudfoundry.org/uaa-go-client/config"
	"code.cloudfoundry.org/uaa-go-client/schema"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("FetchToken", func() {
	var (
		client uaa_go_client.Client
	)

	Context("when configuration is invalid", func() {
		BeforeEach(func() {
			forceUpdate = false
			cfg = &config.Config{
				UaaEndpoint:           "http://not.needed.endpoint",
				MaxNumberOfRetries:    DefaultMaxNumberOfRetries,
				RetryInterval:         DefaultRetryInterval,
				ExpirationBufferInSec: DefaultExpirationBufferTime,
			}
		})
		JustBeforeEach(func() {
			var err error

			clock = fakeclock.NewFakeClock(time.Now())
			logger = lagertest.NewTestLogger("test")
			client, err = uaa_go_client.NewClient(logger, cfg, clock)
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())

		})

		Context("when clientId is missing from config", func() {
			BeforeEach(func() {
				cfg.ClientName = ""
				cfg.ClientSecret = "client-secret"
			})

			It("returns an error", func() {
				_, err := client.FetchToken(forceUpdate)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("OAuth Client ID cannot be empty"))

			})
		})

		Context("when client secret is missing from config", func() {
			BeforeEach(func() {
				cfg.ClientName = "client-name"
				cfg.ClientSecret = ""
			})
			It("returns an error", func() {
				_, err := client.FetchToken(forceUpdate)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("OAuth Client Secret cannot be empty"))

			})
		})
	})

	Context("when configuration is valid", func() {
		BeforeEach(func() {
			forceUpdate = false
			cfg = &config.Config{
				MaxNumberOfRetries:    DefaultMaxNumberOfRetries,
				RetryInterval:         DefaultRetryInterval,
				ExpirationBufferInSec: DefaultExpirationBufferTime,
			}
			server = ghttp.NewServer()

			url, err := url.Parse(server.URL())
			Expect(err).ToNot(HaveOccurred())

			addr := strings.Split(url.Host, ":")

			cfg.UaaEndpoint = "http://" + addr[0] + ":" + addr[1]
			Expect(err).ToNot(HaveOccurred())

			cfg.ClientName = "client-name"
			cfg.ClientSecret = "client-secret"
			clock = fakeclock.NewFakeClock(time.Now())
			logger = lagertest.NewTestLogger("test")

			client, err = uaa_go_client.NewClient(logger, cfg, clock)
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())
		})

		AfterEach(func() {
			server.Close()
		})

		Context("when a new token needs to be fetched from OAuth server", func() {

			Context("when the respose body is malformed", func() {
				It("returns an error and doesn't retry", func() {
					server.AppendHandlers(
						ghttp.RespondWithJSONEncoded(http.StatusOK, "broken garbage response"),
					)

					_, err := client.FetchToken(forceUpdate)
					Expect(err).To(HaveOccurred())
					Expect(server.ReceivedRequests()).Should(HaveLen(1))

					// verifyLogs("test.http-request.*/oauth/token", "test.http-response.*200")
					verifyLogs("test", "test")
				})
			})

			Context("when OAuth server cannot be reached", func() {
				It("retries number of times and finally returns an error", func(done Done) {
					defer close(done)
					cfg.UaaEndpoint = "http://bogus.url:80"
					client, err := uaa_go_client.NewClient(logger, cfg, clock)
					Expect(err).NotTo(HaveOccurred())
					wg := sync.WaitGroup{}
					wg.Add(1)
					go func(wg *sync.WaitGroup) {
						defer GinkgoRecover()
						defer wg.Done()
						_, err := client.FetchToken(forceUpdate)
						Expect(err).To(HaveOccurred())
					}(&wg)

					for i := 0; i < DefaultMaxNumberOfRetries; i++ {
						Eventually(logger).Should(gbytes.Say("fetch-token-from-uaa-start.*bogus.url"))
						Eventually(logger).Should(gbytes.Say("error-fetching-token"))
						clock.WaitForWatcherAndIncrement(DefaultRetryInterval + 10*time.Second)
					}
					wg.Wait()
				}, 5)
			})

			Context("when a non 200 OK is returned", func() {
				Context("when OAuth server returns a 4xx http status code", func() {
					It("returns an error and doesn't retry", func() {
						server.AppendHandlers(
							ghttp.RespondWith(http.StatusBadRequest, "you messed up"),
						)

						_, err := client.FetchToken(forceUpdate)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(Equal("status code: 400, body: you messed up"))
						Expect(server.ReceivedRequests()).Should(HaveLen(1))
					})
				})

				Context("when OAuth server returns a 5xx http status code", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							getOauthHandlerFunc(http.StatusServiceUnavailable, nil),
							getOauthHandlerFunc(http.StatusInternalServerError, nil),
							getOauthHandlerFunc(http.StatusBadGateway, nil),
							getOauthHandlerFunc(http.StatusNotFound, nil),
						)
					})

					It("retries a number of times and finally returns an error", func() {
						verifyFetchWithRetries(client, server, DefaultMaxNumberOfRetries, "status code: 404")
					})
				})

				Context("when OAuth server returns a 3xx http status code", func() {
					It("returns an error and doesn't retry", func() {
						server.AppendHandlers(
							ghttp.RespondWith(http.StatusMovedPermanently, "moved"),
						)

						_, err := client.FetchToken(forceUpdate)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(Equal("status code: 301, body: moved"))
						Expect(server.ReceivedRequests()).Should(HaveLen(1))
					})
				})

				Context("when OAuth server returns a mix of 5xx and 3xx http status codes", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							getOauthHandlerFunc(http.StatusServiceUnavailable, nil),
							getOauthHandlerFunc(http.StatusMovedPermanently, nil),
						)
					})

					It("retries until it hits 3XX status code and returns an error", func() {
						verifyFetchWithRetries(client, server, 2, "status code: 301")
					})
				})
			})

			Context("when OAuth server returns 200 OK", func() {
				It("returns a new token and trace the request response", func() {
					stdout := bytes.NewBuffer([]byte{})
					trace.SetStdout(stdout)
					trace.NewLogger("true")

					responseBody := &schema.Token{
						AccessToken: "the token",
						ExpiresIn:   20,
					}

					server.AppendHandlers(
						getOauthHandlerFunc(http.StatusOK, responseBody),
					)

					token, err := client.FetchToken(forceUpdate)
					Expect(err).NotTo(HaveOccurred())
					Expect(server.ReceivedRequests()).Should(HaveLen(1))
					Expect(token.AccessToken).To(Equal("the token"))
					Expect(token.ExpiresIn).To(Equal(int64(20)))

					r, err := ioutil.ReadAll(stdout)
					Expect(err).NotTo(HaveOccurred())
					log := string(r)
					Expect(log).To(ContainSubstring("REQUEST:"))
					Expect(log).To(ContainSubstring("POST /oauth/token HTTP/1.1"))
					Expect(log).To(ContainSubstring("RESPONSE:"))
					Expect(log).To(ContainSubstring("HTTP/1.1 200 OK"))
				})

				Context("when multiple goroutines fetch a token", func() {
					It("contacts oauth server only once and returns cached token", func() {
						responseBody := &schema.Token{
							AccessToken: "the token",
							ExpiresIn:   3600,
						}

						server.AppendHandlers(
							getOauthHandlerFunc(http.StatusOK, responseBody),
						)
						wg := sync.WaitGroup{}
						for i := 0; i < 2; i++ {
							wg.Add(1)
							go func(wg *sync.WaitGroup) {
								defer GinkgoRecover()
								defer wg.Done()
								token, err := client.FetchToken(forceUpdate)
								Expect(err).NotTo(HaveOccurred())
								Expect(server.ReceivedRequests()).Should(HaveLen(1))
								Expect(token.AccessToken).To(Equal("the token"))
								Expect(token.ExpiresIn).To(Equal(int64(3600)))
							}(&wg)
						}
						wg.Wait()
					})
				})
			})
		})

		Context("when fetching token from Cache", func() {
			Context("when cached token is expired", func() {
				It("returns a new token and logs request response", func() {
					firstResponseBody := &schema.Token{
						AccessToken: "the token",
						ExpiresIn:   3600,
					}
					secondResponseBody := &schema.Token{
						AccessToken: "another token",
						ExpiresIn:   3600,
					}

					server.AppendHandlers(
						getOauthHandlerFunc(http.StatusOK, firstResponseBody),
						getOauthHandlerFunc(http.StatusOK, secondResponseBody),
					)

					token, err := client.FetchToken(forceUpdate)
					Expect(err).NotTo(HaveOccurred())
					Expect(server.ReceivedRequests()).Should(HaveLen(1))
					Expect(token.AccessToken).To(Equal("the token"))
					Expect(token.ExpiresIn).To(Equal(int64(3600)))
					clock.Increment((3600 - DefaultExpirationBufferTime) * time.Second)

					token, err = client.FetchToken(forceUpdate)
					Expect(err).NotTo(HaveOccurred())
					Expect(server.ReceivedRequests()).Should(HaveLen(2))
					Expect(token.AccessToken).To(Equal("another token"))
					Expect(token.ExpiresIn).To(Equal(int64(3600)))
				})
			})

			Context("when a cached token can be used", func() {
				It("returns the cached token", func() {
					firstResponseBody := &schema.Token{
						AccessToken: "the token",
						ExpiresIn:   3600,
					}
					secondResponseBody := &schema.Token{
						AccessToken: "another token",
						ExpiresIn:   3600,
					}

					server.AppendHandlers(
						getOauthHandlerFunc(http.StatusOK, firstResponseBody),
						getOauthHandlerFunc(http.StatusOK, secondResponseBody),
					)

					token, err := client.FetchToken(forceUpdate)
					Expect(err).NotTo(HaveOccurred())
					Expect(server.ReceivedRequests()).Should(HaveLen(1))
					Expect(token.AccessToken).To(Equal("the token"))
					Expect(token.ExpiresIn).To(Equal(int64(3600)))
					clock.Increment(3000 * time.Second)

					token, err = client.FetchToken(forceUpdate)
					Expect(err).NotTo(HaveOccurred())
					Expect(server.ReceivedRequests()).Should(HaveLen(1))
					Expect(token.AccessToken).To(Equal("the token"))
					Expect(token.ExpiresIn).To(Equal(int64(3600)))
				})
			})

			Context("when forcing token refresh", func() {
				It("returns a new token", func() {
					firstResponseBody := &schema.Token{
						AccessToken: "the token",
						ExpiresIn:   3600,
					}
					secondResponseBody := &schema.Token{
						AccessToken: "another token",
						ExpiresIn:   3600,
					}

					server.AppendHandlers(
						getOauthHandlerFunc(http.StatusOK, firstResponseBody),
						getOauthHandlerFunc(http.StatusOK, secondResponseBody),
					)

					token, err := client.FetchToken(forceUpdate)
					Expect(err).NotTo(HaveOccurred())
					Expect(server.ReceivedRequests()).Should(HaveLen(1))
					Expect(token.AccessToken).To(Equal("the token"))
					Expect(token.ExpiresIn).To(Equal(int64(3600)))

					token, err = client.FetchToken(true)
					Expect(err).NotTo(HaveOccurred())
					Expect(server.ReceivedRequests()).Should(HaveLen(2))
					Expect(token.AccessToken).To(Equal("another token"))
					Expect(token.ExpiresIn).To(Equal(int64(3600)))
				})
			})
		})
	})
})
