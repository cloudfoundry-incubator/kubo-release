package uaa_go_client_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"code.cloudfoundry.org/lager/lagertest"

	"code.cloudfoundry.org/uaa-go-client"
	"code.cloudfoundry.org/uaa-go-client/config"

	"code.cloudfoundry.org/clock/fakeclock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Fetch Key", func() {

	var (
		client uaa_go_client.Client
		err    error
		key    string
	)

	Context("FetchKey", func() {
		BeforeEach(func() {
			cfg = &config.Config{
				MaxNumberOfRetries:    DefaultMaxNumberOfRetries,
				RetryInterval:         DefaultRetryInterval,
				RequestTimeout:        DefaultRequestTimeout,
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

		JustBeforeEach(func() {
			key, err = client.FetchKey()
		})

		Context("when UAA is available and responsive", func() {

			Context("and http request succeeds", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", TokenKeyEndpoint),
							ghttp.RespondWith(http.StatusOK, `{}`),
						),
					)
				})
				It("does return an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(key).To(BeEmpty())
				})
			})

			Context("and returns a valid uaa key", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", TokenKeyEndpoint),
							ghttp.RespondWith(http.StatusOK, fmt.Sprintf("{\"alg\":\"alg\", \"value\": \"%s\" }", ValidPemPublicKey)),
						),
					)
				})

				It("returns the key value", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(key).NotTo(BeNil())
					Expect(key).Should(Equal(PemDecodedKey))
				})
			})

			Context("and returns a invalid json uaa key", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", TokenKeyEndpoint),
							ghttp.RespondWith(http.StatusOK, `{"alg":"alg", "value": "ooooppps }`),
						),
					)
				})

				It("returns the error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).Should(Equal(errors.New("unmarshalling error: unexpected EOF")))
					Expect(key).To(BeEmpty())
				})
			})

			Context("and returns a invalid pem key", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", TokenKeyEndpoint),
							ghttp.RespondWith(http.StatusOK, `{"alg":"alg", "value": "not-valid-pem" }`),
						),
					)
				})

				It("returns the error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring(("must be PEM encoded")))
					Expect(key).To(BeEmpty())
				})
			})

			Context("and returns an http error", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", TokenKeyEndpoint),
							ghttp.RespondWith(http.StatusInternalServerError, `{}`),
						),
					)
				})

				It("returns the error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).Should(ContainSubstring("http error"))
					Expect(key).To(BeEmpty())
				})
			})
		})

		Context("when UAA is unavailable", func() {
			BeforeEach(func() {
				cfg.UaaEndpoint = "http://127.0.0.1:1111"
				client, err = uaa_go_client.NewClient(logger, cfg, clock)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns the error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("connection refused"))
				Expect(key).To(BeEmpty())
			})
		})
	})
})
