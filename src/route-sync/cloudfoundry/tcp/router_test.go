package router_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	. "route-sync/cloudfoundry/tcp"
	"route-sync/route"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("routing-api TCP router", func() {
	It("requires credentials and an API endpoint", func() {
		var invalidCreations = []struct {
			creds string
			api   string
		}{
			{"", ""},
			{"foo", ""},
			{"", "bar"},
		}

		for _, t := range invalidCreations {
			Context(fmt.Sprintf("creds: %s, api: %s", t.creds, t.api), func() {
				router, err := NewRoutingApi(t.creds, t.api)
				Expect(router).To(BeNil())
				Expect(err).NotTo(BeNil())
			})
		}

		router, err := NewRoutingApi("foo", "foo")
		Expect(router).NotTo(BeNil())
		Expect(err).To(BeNil())
	})

	It("posts the UAA token during a request", func() {
		requestChan := make(chan *http.Request)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			go func() {
				requestChan <- r
			}()
		}))
		defer ts.Close()

		router, _ := NewRoutingApi("foobar", ts.URL)

		router.RouterGroups()
		req := <-requestChan
		Expect(req.Header).To(HaveKeyWithValue("Authorization", []string{"bearer foobar"}))

		router.CreateRoutes(RouterGroup{}, []route.TCP{route.TCP{}})
		req = <-requestChan
		Expect(req.Header).To(HaveKeyWithValue("Authorization", []string{"bearer foobar"}))
	})
})
