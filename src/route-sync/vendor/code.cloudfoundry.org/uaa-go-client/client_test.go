package uaa_go_client_test

import (
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/uaa-go-client"
	"code.cloudfoundry.org/uaa-go-client/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("UAA Client", func() {
	Context("non-TLS client", func() {

		BeforeEach(func() {
			forceUpdate = false
			cfg = &config.Config{
				MaxNumberOfRetries:    DefaultMaxNumberOfRetries,
				RetryInterval:         DefaultRetryInterval,
				ExpirationBufferInSec: DefaultExpirationBufferTime,
				RequestTimeout:        DefaultRequestTimeout,
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
		})

		AfterEach(func() {
			server.Close()
		})

		Describe("uaa_go_client.NewClient", func() {
			Context("when all values are valid", func() {
				It("returns a token fetcher instance", func() {
					client, err := uaa_go_client.NewClient(logger, cfg, clock)
					Expect(err).NotTo(HaveOccurred())
					Expect(client).NotTo(BeNil())
				})
			})

			Context("when values are invalid", func() {
				Context("when oauth config is nil", func() {
					It("returns error", func() {
						client, err := uaa_go_client.NewClient(logger, nil, clock)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Configuration cannot be nil"))
						Expect(client).To(BeNil())
					})
				})

				Context("when oauth config client id is empty", func() {
					It("creates new client", func() {
						config := &config.Config{
							UaaEndpoint:  "http://some.url:80",
							ClientName:   "",
							ClientSecret: "client-secret",
						}
						client, err := uaa_go_client.NewClient(logger, config, clock)
						Expect(err).ToNot(HaveOccurred())
						Expect(client).ToNot(BeNil())
					})
				})

				Context("when oauth config client secret is empty", func() {
					It("creates a new client", func() {
						config := &config.Config{
							UaaEndpoint:  "http://some.url:80",
							ClientName:   "client-name",
							ClientSecret: "",
						}
						client, err := uaa_go_client.NewClient(logger, config, clock)
						Expect(err).ToNot(HaveOccurred())
						Expect(client).ToNot(BeNil())
					})
				})

				Context("when oauth config tokenendpoint is empty", func() {
					It("returns error", func() {
						config := &config.Config{
							UaaEndpoint:  "",
							ClientName:   "client-name",
							ClientSecret: "client-secret",
						}
						client, err := uaa_go_client.NewClient(logger, config, clock)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("UAA endpoint cannot be empty"))
						Expect(client).To(BeNil())
					})
				})

				Context("when token fetcher config's max number of retries is zero", func() {
					It("creates the client", func() {
						config := &config.Config{
							UaaEndpoint:           "http://some.url:80",
							MaxNumberOfRetries:    0,
							RetryInterval:         2 * time.Second,
							ExpirationBufferInSec: 30,
							ClientName:            "client-name",
							ClientSecret:          "client-secret",
						}
						client, err := uaa_go_client.NewClient(logger, config, clock)
						Expect(err).NotTo(HaveOccurred())
						Expect(client).NotTo(BeNil())
					})
				})

				Context("when token fetcher config's expiration buffer time is negative", func() {
					It("sets the expiration buffer time to the default value", func() {
						config := &config.Config{
							MaxNumberOfRetries:    3,
							RetryInterval:         2 * time.Second,
							ExpirationBufferInSec: -1,
							UaaEndpoint:           "http://some.url:80",
							ClientName:            "client-name",
							ClientSecret:          "client-secret",
							RequestTimeout:        DefaultRequestTimeout,
						}
						client, err := uaa_go_client.NewClient(logger, config, clock)
						Expect(err).NotTo(HaveOccurred())
						Expect(client).NotTo(BeNil())
					})
				})
			})
		})
	})

	Describe("Decode token", func() {
		var (
			uaaClient    uaa_go_client.Client
			publicKeyPEM []byte
			privateKey   *rsa.PrivateKey
		)

		BeforeEach(func() {
			var err error
			var publicKey *rsa.PublicKey
			privateKey, publicKey, err = generateRSAKeyPair()
			Expect(err).NotTo(HaveOccurred())
			publicKeyPEM, err = publicKeyToPEM(publicKey)
			Expect(err).NotTo(HaveOccurred())

			cfg = &config.Config{
				MaxNumberOfRetries:    DefaultMaxNumberOfRetries,
				RetryInterval:         DefaultRetryInterval,
				ExpirationBufferInSec: DefaultExpirationBufferTime,
				RequestTimeout:        DefaultRequestTimeout,
			}

			server = ghttp.NewServer()

			url, err := url.Parse(server.URL())
			Expect(err).ToNot(HaveOccurred())
			cfg.UaaEndpoint = "http://" + url.Host

			var uaaResponseStruct = struct {
				Alg   string `json:"alg"`
				Value string `json:"value"`
			}{"alg", string(publicKeyPEM)}
			server.AppendHandlers(
				ghttp.RespondWithJSONEncoded(
					http.StatusOK,
					uaaResponseStruct,
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", OpenIDConfigEndpoint),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf("{\"issuer\":\"https://uaa.domain.com\"}")),
				),
			)

			cfg.ClientName = "client-name"
			cfg.ClientSecret = "client-secret"
			clock = fakeclock.NewFakeClock(time.Now())
			logger = lagertest.NewTestLogger("test")

			uaaClient, err = uaa_go_client.NewClient(logger, cfg, clock)
			Expect(err).NotTo(HaveOccurred())
			Expect(uaaClient).NotTo(BeNil())
		})

		AfterEach(func() {
			server.Close()
		})

		Context("when the token has been signed with the correct private key", func() {
			It("succeeds", func() {
				validToken, err := makeValidToken(privateKey)
				Expect(err).NotTo(HaveOccurred())
				err = uaaClient.DecodeToken(validToken, "some.scope")
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the token algorithm doesn't match the key type", func() {
			It("fails", func() {
				spoofedToken, err := makeSpoofedToken(publicKeyPEM)
				Expect(err).NotTo(HaveOccurred())
				err = uaaClient.DecodeToken(spoofedToken, "some.scope")
				Expect(err).To(MatchError("invalid signing method"))
			})
		})
	})

	Describe("Fetch Issuer Metadata", func() {
		var (
			uaaClient    uaa_go_client.Client
			publicKeyPEM []byte
			privateKey   *rsa.PrivateKey
		)
		BeforeEach(func() {
			var err error
			var publicKey *rsa.PublicKey
			privateKey, publicKey, err = generateRSAKeyPair()
			Expect(err).NotTo(HaveOccurred())
			publicKeyPEM, err = publicKeyToPEM(publicKey)
			Expect(err).NotTo(HaveOccurred())

			cfg = &config.Config{
				MaxNumberOfRetries:    DefaultMaxNumberOfRetries,
				RetryInterval:         DefaultRetryInterval,
				ExpirationBufferInSec: DefaultExpirationBufferTime,
				RequestTimeout:        DefaultRequestTimeout,
			}

			server = ghttp.NewServer()

			url, err := url.Parse(server.URL())
			Expect(err).ToNot(HaveOccurred())
			cfg.UaaEndpoint = "http://" + url.Host

			cfg.ClientName = "client-name"
			cfg.ClientSecret = "client-secret"
			clock = fakeclock.NewFakeClock(time.Now())
			logger = lagertest.NewTestLogger("test")

			uaaClient, err = uaa_go_client.NewClient(logger, cfg, clock)
			Expect(err).NotTo(HaveOccurred())
			Expect(uaaClient).NotTo(BeNil())
		})

		AfterEach(func() {
			server.Close()
		})

		Context("when fetch issuer is already invoked", func() {
			var privateKey *rsa.PrivateKey
			BeforeEach(func() {

				var err error
				var publicKey *rsa.PublicKey
				privateKey, publicKey, err = generateRSAKeyPair()
				Expect(err).NotTo(HaveOccurred())
				publicKeyPEM, err = publicKeyToPEM(publicKey)
				Expect(err).NotTo(HaveOccurred())
				var uaaResponseStruct = struct {
					Alg   string `json:"alg"`
					Value string `json:"value"`
				}{"alg", string(publicKeyPEM)}
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", OpenIDConfigEndpoint),
						ghttp.RespondWith(http.StatusOK, fmt.Sprintf("{\"issuer\":\"https://uaa.domain.com\"}")),
					),
					ghttp.RespondWithJSONEncoded(
						http.StatusOK,
						uaaResponseStruct,
					),
				)
				_, err = uaaClient.FetchIssuer()
				Expect(err).ToNot(HaveOccurred())
			})
			It("does not call issuer endpoint again for decoding tokens", func() {
				validToken, err := makeValidToken(privateKey)
				Expect(err).NotTo(HaveOccurred())

				err = uaaClient.DecodeToken(validToken, "some.scope")
				Expect(err).ToNot(HaveOccurred())
				Expect(len(server.ReceivedRequests())).To(Equal(2))
			})
		})

		Context("when UAA server responds with valid metadata structure", func() {
			BeforeEach(func() {
				uaaResponse := `{"issuer":"https://uaa.domain.com"}`

				server.AppendHandlers(
					ghttp.RespondWith(
						http.StatusOK,
						uaaResponse,
					),
				)
			})
			It("successfully unmarshall the metadata info", func() {

				issuer, err := uaaClient.FetchIssuer()
				Expect(err).ToNot(HaveOccurred())
				Expect(issuer).To(Equal("https://uaa.domain.com"))
			})
		})
		Context("when UAA server responds with invalid metadata structure", func() {
			BeforeEach(func() {
				uaaResponse := `{"https://uaa.domain.com"}`

				server.AppendHandlers(
					ghttp.RespondWith(
						http.StatusOK,
						uaaResponse,
					),
				)
			})
			It("returns an error", func() {

				issuer, err := uaaClient.FetchIssuer()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid character"))
				Expect(issuer).To(Equal(""))
			})
		})
	})
	Context("secure (TLS) client", func() {

		var (
			tlsServer   *http.Server
			tlsListener net.Listener
		)
		BeforeEach(func() {
			forceUpdate = false
			cfg = &config.Config{
				MaxNumberOfRetries:    DefaultMaxNumberOfRetries,
				RetryInterval:         DefaultRetryInterval,
				ExpirationBufferInSec: DefaultExpirationBufferTime,
				RequestTimeout:        DefaultRequestTimeout,
			}

			listener, err := net.Listen("tcp", "127.0.0.1:0")
			addr := strings.Split(listener.Addr().String(), ":")

			cfg.UaaEndpoint = "https://" + addr[0] + ":" + addr[1]
			Expect(err).NotTo(HaveOccurred())

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(fmt.Sprintf("{\"alg\":\"alg\", \"value\": \"%s\" }", ValidPemPublicKey)))
			})

			tlsListener = newTlsListener(listener)
			tlsServer = &http.Server{Handler: handler}

			go func() {
				err := tlsServer.Serve(tlsListener)
				Expect(err).ToNot(HaveOccurred())
			}()

			Expect(err).ToNot(HaveOccurred())

			cfg.ClientName = "client-name"
			cfg.ClientSecret = "client-secret"

			clock = fakeclock.NewFakeClock(time.Now())
			logger = lagertest.NewTestLogger("test")
		})

		Context("when CA cert provided", func() {
			var (
				tlsClient uaa_go_client.Client
			)

			BeforeEach(func() {
				caCertPath, err := filepath.Abs(path.Join("fixtures", "ca.pem"))
				Expect(err).ToNot(HaveOccurred())

				cfg.CACerts = caCertPath
				cfg.MaxNumberOfRetries = 0
				tlsClient, err = uaa_go_client.NewClient(logger, cfg, clock)
				Expect(err).ToNot(HaveOccurred())
				Expect(tlsClient).ToNot(BeNil())
			})

			It("can make uaa request with cert", func() {
				_, err := tlsClient.FetchToken(true)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when secure uaa client skips verify", func() {
			var (
				tlsClient uaa_go_client.Client
			)

			BeforeEach(func() {
				cfg.SkipVerification = true
				var err error
				tlsClient, err = uaa_go_client.NewClient(logger, cfg, clock)
				Expect(err).ToNot(HaveOccurred())
				Expect(tlsClient).ToNot(BeNil())
			})

			It("logs fetching token", func() {
				_, err := tlsClient.FetchToken(true)
				Expect(err).ToNot(HaveOccurred())
				Expect(logger).To(gbytes.Say("uaa-client"))
				Expect(logger).To(gbytes.Say("started-fetching-token"))
				Expect(logger).To(gbytes.Say(cfg.UaaEndpoint))
				Expect(logger).To(gbytes.Say("successfully-fetched-token"))
			})

			It("logs fetching key", func() {
				_, err := tlsClient.FetchKey()
				Expect(err).ToNot(HaveOccurred())
				Expect(logger).To(gbytes.Say("uaa-client"))
				Expect(logger).To(gbytes.Say("fetch-key-starting"))
				Expect(logger).To(gbytes.Say(cfg.UaaEndpoint))
				Expect(logger).To(gbytes.Say("fetch-key-successful"))
			})
		})
	})
})

func newTlsListener(listener net.Listener) net.Listener {
	public := "fixtures/server.pem"
	private := "fixtures/server.key"
	cert, err := tls.LoadX509KeyPair(public, private)
	Expect(err).ToNot(HaveOccurred())

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		CipherSuites: []uint16{tls.TLS_RSA_WITH_AES_256_CBC_SHA},
	}

	return tls.NewListener(listener, tlsConfig)
}

const (
	pemHeaderPrivateKey = "RSA PRIVATE KEY"
	pemHeaderPublicKey  = "PUBLIC KEY"
)

func generateRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	private, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, nil, err
	}
	public := private.Public().(*rsa.PublicKey)
	return private, public, nil
}

func encodePEM(keyBytes []byte, keyType string) []byte {
	block := &pem.Block{
		Type:  keyType,
		Bytes: keyBytes,
	}

	return pem.EncodeToMemory(block)
}

// PrivateKeyToPEM serializes an RSA Private key into PEM format.
func privateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	keyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	return encodePEM(keyBytes, pemHeaderPrivateKey)
}

// PublicKeyToPEM serializes an RSA Public key into PEM format.
func publicKeyToPEM(publicKey *rsa.PublicKey) ([]byte, error) {
	keyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return []byte{}, err
	}

	return encodePEM(keyBytes, pemHeaderPublicKey), nil
}

func jwtHeader(alg, kid string) string {
	return fmt.Sprintf(`{ "alg": "%s", "kid": "%s", "typ": "JWT" }`, alg, kid)
}

var tokenEncoding = base64.RawURLEncoding

func signWithHS256(signingString string, key string) string {
	hasher := hmac.New(sha256.New, []byte(key))
	hasher.Write([]byte(signingString))
	return tokenEncoding.EncodeToString(hasher.Sum(nil))
}

func signWithRS256(signingString string, privateKey *rsa.PrivateKey) (string, error) {
	hasher := crypto.SHA256.New()
	hasher.Write([]byte(signingString))

	sigBytes, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hasher.Sum(nil))
	if err != nil {
		return "", err
	}
	return tokenEncoding.EncodeToString(sigBytes), nil
}

const tokenPayload = `{
  "scope": [
    "some.scope"
  ],
  "iat": 1481253086,
  "exp": 2491253686,
	"iss": "https://uaa.domain.com"
}`

func makeValidToken(privateKey *rsa.PrivateKey) (string, error) {
	header := jwtHeader("RS256", "some-key-id")
	signingString := fmt.Sprintf("%s.%s",
		tokenEncoding.EncodeToString([]byte(header)),
		tokenEncoding.EncodeToString([]byte(tokenPayload)),
	)
	signature, err := signWithRS256(signingString, privateKey)
	if err != nil {
		return "", err
	}
	fullToken := fmt.Sprintf("bearer %s.%s", signingString, signature)
	return fullToken, nil
}

func makeSpoofedToken(publicKeyPEM []byte) (string, error) {
	header := jwtHeader("HS256", "some-key-id")
	signingString := fmt.Sprintf("%s.%s",
		tokenEncoding.EncodeToString([]byte(header)),
		tokenEncoding.EncodeToString([]byte(tokenPayload)),
	)
	signature := signWithHS256(signingString, string(publicKeyPEM))
	fullToken := fmt.Sprintf("bearer %s.%s", signingString, signature)
	return fullToken, nil
}
