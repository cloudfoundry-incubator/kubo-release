package uaa_go_client_test

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/uaa-go-client"
	"code.cloudfoundry.org/uaa-go-client/config"
	"code.cloudfoundry.org/uaa-go-client/schema"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("RegisterOauthClient", func() {
	var (
		client uaa_go_client.Client
	)

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

	Context("when OAuth server returns 200 OK", func() {
		It("does not return an error", func() {
			accessToken := &schema.Token{
				AccessToken: "the token",
				ExpiresIn:   20,
			}

			oauthClient := &schema.OauthClient{
				ClientId:             "clientId",
				Name:                 "the new client",
				ClientSecret:         "secret",
				Scope:                []string{"uaa.none"},
				ResourceIds:          []string{"none"},
				Authorities:          []string{"openid"},
				AuthorizedGrantTypes: []string{"client_credentials"},
				AccessTokenValidity:  10000,
				RedirectUri:          []string{"http://example.com"},
			}

			server.AppendHandlers(
				getOauthHandlerFunc(http.StatusOK, accessToken),
				getRegisterOauthClientHandlerFunc(http.StatusOK, accessToken, oauthClient),
			)

			receivedOauthClient, err := client.RegisterOauthClient(oauthClient)
			Expect(err).NotTo(HaveOccurred())
			Expect(server.ReceivedRequests()).Should(HaveLen(2))
			Expect(receivedOauthClient).NotTo(BeNil())
			Expect(receivedOauthClient.ClientId).To(Equal("clientId"))
		})
	})

	Context("when OAuth server returns 409 Conflict", func() {
		It("returns ErrClientAlreadyExists error", func() {
			accessToken := &schema.Token{
				AccessToken: "the token",
				ExpiresIn:   20,
			}

			oauthClient := &schema.OauthClient{
				ClientId:             "clientId",
				Name:                 "the new client",
				ClientSecret:         "secret",
				Scope:                []string{"uaa.none"},
				ResourceIds:          []string{"none"},
				Authorities:          []string{"openid"},
				AuthorizedGrantTypes: []string{"client_credentials"},
				AccessTokenValidity:  10000,
				RedirectUri:          []string{"http://example.com"},
			}

			server.AppendHandlers(
				getOauthHandlerFunc(http.StatusOK, accessToken),
				getRegisterOauthClientHandlerFunc(http.StatusConflict, accessToken, oauthClient),
			)

			receivedOauthClient, err := client.RegisterOauthClient(oauthClient)
			Expect(err).To(Equal(uaa_go_client.ErrClientAlreadyExists))
			Expect(server.ReceivedRequests()).Should(HaveLen(2))
			Expect(receivedOauthClient).To(BeNil())
		})
	})
})
