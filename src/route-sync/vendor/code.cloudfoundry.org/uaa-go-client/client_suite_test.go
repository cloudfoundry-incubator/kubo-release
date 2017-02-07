package uaa_go_client_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
	"code.cloudfoundry.org/uaa-go-client"
	"code.cloudfoundry.org/uaa-go-client/config"
	"code.cloudfoundry.org/uaa-go-client/schema"

	"code.cloudfoundry.org/lager"
	"encoding/json"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/ghttp"
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client Suite")
}

const (
	TokenKeyEndpoint            = "/token_key"
	DefaultMaxNumberOfRetries   = 3
	DefaultRetryInterval        = 15 * time.Second
	DefaultExpirationBufferTime = 30
	ValidPemPublicKey           = `-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDHFr+KICms+tuT1OXJwhCUmR2d\nKVy7psa8xzElSyzqx7oJyfJ1JZyOzToj9T5SfTIq396agbHJWVfYphNahvZ/7uMX\nqHxf+ZH9BL1gk9Y6kCnbM5R60gfwjyW1/dQPjOzn9N394zd2FJoFHwdq9Qs0wBug\nspULZVNRxq7veq/fzwIDAQAB\n-----END PUBLIC KEY-----`
	InvalidPemPublicKey         = `-----BEGIN PUBLIC KEY-----\nMJGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDHFr+KICms+tuT1OXJwhCUmR2d\nKVy7psa8xzElSyzqx7oJyfJ1JZyOzToj9T5SfTIq396agbHJWVfYphNahvZ/7uMX\nqHxf+ZH9BL1gk9Y6kCnbM5R60gfwjyW1/dQPjOzn9N394zd2FJoFHwdq9Qs0wBug\nspULZVNRxq7veq/fzwIDAQAB\n-----END PUBLIC KEY-----`
	PemDecodedKey               = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDHFr+KICms+tuT1OXJwhCUmR2d
KVy7psa8xzElSyzqx7oJyfJ1JZyOzToj9T5SfTIq396agbHJWVfYphNahvZ/7uMX
qHxf+ZH9BL1gk9Y6kCnbM5R60gfwjyW1/dQPjOzn9N394zd2FJoFHwdq9Qs0wBug
spULZVNRxq7veq/fzwIDAQAB
-----END PUBLIC KEY-----`
)

var (
	logger      lager.Logger
	forceUpdate bool
	server      *ghttp.Server
	clock       *fakeclock.FakeClock
	cfg         *config.Config
)

var verifyBody = func(expectedBody string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		Expect(err).ToNot(HaveOccurred())

		defer r.Body.Close()
		Expect(string(body)).To(Equal(expectedBody))
	}
}

var verifyLogs = func(reqMessage, resMessage string) {
	Expect(logger).To(gbytes.Say(reqMessage))
	Expect(logger).To(gbytes.Say(resMessage))
}

var getOauthHandlerFunc = func(status int, token *schema.Token) http.HandlerFunc {
	return ghttp.CombineHandlers(
		ghttp.VerifyRequest("POST", "/oauth/token"),
		ghttp.VerifyBasicAuth("client-name", "client-secret"),
		ghttp.VerifyContentType("application/x-www-form-urlencoded; charset=UTF-8"),
		ghttp.VerifyHeader(http.Header{
			"Accept": []string{"application/json; charset=utf-8"},
		}),
		verifyBody("grant_type=client_credentials"),
		ghttp.RespondWithJSONEncoded(status, token),
	)
}

var getSuccessKeyFetchHandler = func(key string) http.HandlerFunc {
	return ghttp.CombineHandlers(
		ghttp.VerifyRequest("GET", TokenKeyEndpoint),
		ghttp.RespondWith(http.StatusOK, fmt.Sprintf("{\"alg\":\"alg\", \"value\": \"%s\" }", key)),
	)
}

var verifyFetchWithRetries = func(client uaa_go_client.Client, server *ghttp.Server, numRetries int, expectedErrorMsg string) {
	var err error
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer GinkgoRecover()
		defer wg.Done()
		_, err = client.FetchToken(forceUpdate)
		Expect(err).To(HaveOccurred())
	}(&wg)

	for i := 0; i < numRetries; i++ {
		Eventually(server.ReceivedRequests, 7*time.Second, 1*time.Second).Should(HaveLen(i + 1))
		clock.Increment(DefaultRetryInterval + 10*time.Second)
	}

	wg.Wait()

	Expect(err.Error()).To(ContainSubstring(expectedErrorMsg))
}

var getRegisterOauthClientHandlerFunc = func(status int, token *schema.Token, oauthClient *schema.OauthClient) http.HandlerFunc {
	oauthClientString, err := json.Marshal(oauthClient)
	var responseBody string
	if status == http.StatusOK {
		responseBody = string(oauthClientString)
	} else {
		responseBody = ""
	}

	Expect(err).ToNot(HaveOccurred())

	return ghttp.CombineHandlers(
		ghttp.VerifyRequest("POST", "/oauth/clients"),
		ghttp.VerifyContentType("application/json; charset=UTF-8"),
		ghttp.VerifyHeader(http.Header{
			"Accept":        []string{"application/json; charset=utf-8"},
			"Authorization": []string{"bearer " + token.AccessToken},
		}),
		verifyBody(string(oauthClientString)),
		ghttp.RespondWith(status, responseBody),
	)
}
